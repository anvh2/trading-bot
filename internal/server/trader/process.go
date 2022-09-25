package trader

import (
	"context"
	"errors"
	"fmt"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/helpers"
	"github.com/anvh2/trading-bot/internal/indicator"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	positionCount = 0
)

func (s *Server) ProcessTrading(ctx context.Context, msg interface{}) error {
	message := msg.(*models.Oscillator)

	if err := validateTradingMessage(message); err != nil {
		return err
	}

	if positionCount > 2 {
		return errors.New("trading: can't trade any more")
	}

	openPositions, err := s.binance.ListPositionRisk(ctx, message.Symbol)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get positions", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	if checkExistPosition(openPositions) {
		return nil
	}

	openOrders, err := s.binance.ListOpenOrders(ctx, message.Symbol)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get orders", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	if len(openOrders) > 0 {
		return nil
	}

	symbolPrice, err := s.binance.GetCurrentPrice(ctx, message.Symbol)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get current symbol price", zap.Any("symbol", message.Symbol), zap.Error(err))
		return err
	}

	candles, err := s.binance.ListCandlesticks(ctx, message.Symbol, "1h", 2)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get candles", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	orders, err := s.makeOrders(message.Symbol, symbolPrice.Price, candles, message.Stoch["1h"])
	if err != nil {
		s.logger.Info("[ProcessTrading] failed to make orders", zap.String("price", symbolPrice.Price), zap.Any("candles", candles), zap.Any("stoch", message.Stoch["1h"]), zap.Error(err))
		return err
	}

	s.logger.Info("[ProcessTrading] make orders success", zap.String("symbol", message.Symbol), zap.String("price", symbolPrice.Price),
		zap.Any("candles", candles), zap.Any("stoch", message.Stoch["1h"]), zap.Any("orders", orders))

	resp, err := s.binance.CreateOrders(ctx, orders)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to create orders", zap.Any("orders", orders), zap.Error(err))
		return err
	}

	positionCount++

	s.supbot.PushNotify(ctx, viper.GetInt64("notify.channels.orders_creation"), fmt.Sprintf("Create Orders Success: %s", message.Symbol))
	s.logger.Info("[ProcessTrading] create order success", zap.Any("resp", resp))

	return nil
}

func validateTradingMessage(message *models.Oscillator) error {
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
			Quantity:         helpers.AlignQuantityToString(calculateQuantity(entryPrice, 20), lotFilter.StepSize),
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
			Quantity:         helpers.AlignQuantityToString(calculateQuantity(entryPrice*1.03, 30), lotFilter.StepSize),
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
			Quantity:         helpers.AlignQuantityToString(calculateQuantity(entryPrice*1.03*1.03, 40), lotFilter.StepSize),
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
			Quantity:         helpers.AlignQuantityToString(calculateStopQuantity(entryPrice, 60), lotFilter.StepSize),
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
			Quantity:         helpers.AlignQuantityToString(calculateStopQuantity(entryPrice, 60), lotFilter.StepSize),
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
