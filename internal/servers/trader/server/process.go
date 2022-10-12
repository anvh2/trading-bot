package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/helpers"
	"github.com/anvh2/trading-bot/internal/indicator"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Server) Process(ctx context.Context, data interface{}) error {
	message := &models.Oscillator{}

	if err := json.Unmarshal([]byte(fmt.Sprint(data)), message); err != nil {
		s.logger.Error("[Process] failed to unmarshal message", zap.Error(err))
		return err
	}

	if err := validateMessage(message); err != nil {
		return err
	}

	openPositions, err := s.binance.ListPositionRisk(ctx, message.Symbol)
	if err != nil {
		s.logger.Error("[Process] failed to get positions", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	if checkExistPosition(openPositions) {
		s.logger.Info("[Process] position existed", zap.String("symbol", message.Symbol), zap.Any("openPositions", openPositions))
		return nil
	}

	openOrders, err := s.binance.ListOpenOrders(ctx, message.Symbol)
	if err != nil {
		s.logger.Error("[Process] failed to get orders", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	if len(openOrders) > 0 {
		s.logger.Info("[Process] order existed", zap.String("symbol", message.Symbol), zap.Any("orders", openOrders))
		return nil
	}

	symbolPrice, err := s.binance.GetCurrentPrice(ctx, message.Symbol)
	if err != nil {
		s.logger.Error("[Process] failed to get current symbol price", zap.Any("symbol", message.Symbol), zap.Error(err))
		return err
	}

	candles, err := s.binance.ListCandlesticks(ctx, message.Symbol, "1h", 2)
	if err != nil {
		s.logger.Error("[Process] failed to get candles", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	orders, err := s.makeOrders(message.Symbol, symbolPrice.Price, candles, message.Stoch["1h"])
	if err != nil {
		s.logger.Info("[Process] failed to make orders", zap.String("price", symbolPrice.Price), zap.Any("candles", candles), zap.Any("stoch", message.Stoch["1h"]), zap.Error(err))
		return err
	}

	s.logger.Info("[Process] make orders success", zap.String("symbol", message.Symbol), zap.String("price", symbolPrice.Price),
		zap.Any("candles", candles), zap.Any("stoch", message.Stoch["1h"]), zap.Any("orders", orders))

	resp, err := s.binance.CreateOrders(ctx, orders)
	if err != nil {
		s.logger.Error("[Process] failed to create orders", zap.Any("orders", orders), zap.Error(err))
		return err
	}

	channel := cast.ToString(viper.GetInt64("notify.channels.orders_creation"))
	notifyMsg := fmt.Sprintf("Create orders with %s side for %s success.", helpers.ResolvePositionSide(message.GetRSI()), message.Symbol)
	_, err = s.notifier.Push(ctx, &notifier.PushRequest{Channel: channel, Message: notifyMsg})
	if err != nil {
		s.logger.Error("[Process] failed to push notification", zap.Error(err))
		return err
	}

	s.logger.Info("[Process] create order success", zap.Any("resp", resp))
	return nil
}

func validateMessage(message *models.Oscillator) error {
	if message == nil {
		return errors.New("trading: message invalid")
	}
	return nil
}

func checkExistPosition(positions []*futures.PositionRisk) bool {
	for _, pos := range positions {
		if pos.EntryPrice != "0.0" {
			return true
		}
	}
	return false
}

func (s *Server) makeOrders(symbol string, currentPrice string, candles []*binance.Kline, stoch *models.Stoch) ([]*models.Order, error) {
	if stoch == nil {
		return nil, errors.New("orders: empty stoch")
	}

	if !indicator.WithinRangeBound(stoch, indicator.RangeBoundReadyTrade) {
		return nil, errors.New("orders: indicator not ready to trade")
	}

	var (
		sideType  futures.SideType
		closeSide futures.SideType
	)

	positionSide, err := indicator.ResolvePositionSide(stoch, indicator.RangeBoundReadyTrade)
	if err != nil {
		return nil, err
	}

	switch positionSide {
	case futures.PositionSideTypeShort:
		sideType = futures.SideTypeSell
		closeSide = futures.SideTypeBuy
	case futures.PositionSideTypeLong:
		sideType = futures.SideTypeBuy
		closeSide = futures.SideTypeSell
	}

	if len(candles) < 2 {
		return nil, errors.New("orders: len of candles not enough")
	}

	const (
		longGain  = 1.7
		longLoss  = 0.5
		shortGain = 0.5
		shortLoss = 1.5
	)

	var (
		entryPrice      float64
		takeProfitPrice float64
		stopLossPrice   float64
	)

	switch positionSide {
	case futures.PositionSideTypeShort:
		entryPrice = helpers.AddFloat(candles[0].High, candles[1].High, currentPrice) / 3.0

		current := helpers.StringToFloat(currentPrice)
		if entryPrice < current {
			entryPrice = current * 1.01
		}

		takeProfitPrice = entryPrice * shortGain
		stopLossPrice = entryPrice * shortLoss

	case futures.PositionSideTypeLong:
		entryPrice = helpers.AddFloat(candles[0].Low, candles[1].Low, currentPrice) / 3.0

		current := helpers.StringToFloat(currentPrice)
		if entryPrice > current {
			entryPrice = current * 0.99
		}

		takeProfitPrice = entryPrice * longGain
		stopLossPrice = entryPrice * shortGain

	}

	exchange, err := s.exchange.Get(symbol)
	if err != nil {
		return nil, err
	}

	priceFilter, err := exchange.GetPriceFilter()
	if err != nil {
		return nil, err
	}

	lotFilter, err := exchange.GetLotSizeFilter()
	if err != nil {
		return nil, err
	}

	orders := []*models.Order{
		// DCA the first three
		{
			Symbol:           symbol,
			Side:             sideType,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.AlignQuantityToString(calculateQuantity(entryPrice, 30), lotFilter.StepSize),
			Price:            helpers.AlignPriceToString(entryPrice, priceFilter.TickSize),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		{
			Symbol:           symbol,
			Side:             sideType,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.AlignQuantityToString(calculateQuantity(entryPrice*1.03, 40), lotFilter.StepSize),
			Price:            helpers.AlignPriceToString(entryPrice*1.03, priceFilter.TickSize),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		{
			Symbol:           symbol,
			Side:             sideType,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.AlignQuantityToString(calculateQuantity(entryPrice*1.03*1.03, 50), lotFilter.StepSize),
			Price:            helpers.AlignPriceToString(entryPrice*1.03*1.03, priceFilter.TickSize),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		// take profile
		{
			Symbol:           symbol,
			Side:             closeSide,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeTakeProfitMarket,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.AlignQuantityToString(calculateStopQuantity(entryPrice, 120), lotFilter.StepSize),
			StopPrice:        helpers.AlignPriceToString(takeProfitPrice, priceFilter.TickSize),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		// stop loss
		{
			Symbol:           symbol,
			Side:             closeSide,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeStopMarket,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.AlignQuantityToString(calculateStopQuantity(entryPrice, 120), lotFilter.StepSize),
			StopPrice:        helpers.AlignPriceToString(stopLossPrice, priceFilter.TickSize),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
	}
	return orders, nil
}

func calculateQuantity(price, amount float64) float64 {
	return amount / price
}

func calculateStopQuantity(price float64, totalAmount float64) float64 {
	return totalAmount / price
}
