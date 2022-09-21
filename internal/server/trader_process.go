package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/helpers"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

var (
	positionCount = 0
)

func (s *Server) ProcessTrading(ctx context.Context, message *models.Oscillator) error {
	if err := validateTradingMessage(message); err != nil {
		return err
	}

	if positionCount > 2 {
		return nil
	}

	openPositions, err := s.binance.NewGetPositionRiskService().Symbol(message.Symbol).Do(ctx)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get positions", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	if checkExistPosition(openPositions) {
		return nil
	}

	openOrders, err := s.binance.NewListOpenOrdersService().Symbol(message.Symbol).Do(ctx)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get orders", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	if len(openOrders) > 0 {
		return nil
	}

	symbolPrice, err := s.futures.GetCurrentPrice(ctx, message.Symbol)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get current symbol price", zap.Any("symbol", message.Symbol), zap.Error(err))
		return err
	}

	candles, err := s.binance.NewKlinesService().Symbol(message.Symbol).Interval("1h").Limit(2).Do(ctx)
	if err != nil {
		s.logger.Error("[ProcessTrading] failed to get candles", zap.String("symbol", message.Symbol), zap.Error(err))
		return err
	}

	orders, err := makeOrders(message.Symbol, symbolPrice.Price, candles, message.Stoch["1h"])
	if err != nil {
		s.logger.Info("[ProcessTrading] failed to make orders", zap.String("price", symbolPrice.Price), zap.Any("candles", candles), zap.Any("stoch", message.Stoch["1h"]), zap.Error(err))
		return err
	}

	s.logger.Info("[ProcessTrading] make orders success", zap.Any("orders", orders))

	// resp, err := s.futures.CreateOrders(ctx, orders)
	// if err != nil {
	// 	s.logger.Error("[ProcessTrading] failed to create orders", zap.Any("orders", orders), zap.Error(err))
	// 	return err
	// }

	positionCount++

	s.supbot.PushNotify(ctx, config.TelegramUserId, fmt.Sprintf("Create Orders Success: %s", message.Symbol))
	// s.logger.Info("[ProcessTrading] create order success", zap.Any("resp", resp))

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

func makeOrders(symbol string, currentPrice string, candles []*futures.Kline, stoch *models.Stoch) ([]*models.Order, error) {
	if stoch == nil {
		return nil, errors.New("orders: empty stoch")
	}

	if stoch.RSI < 80 || stoch.RSI > 15 {
		return nil, errors.New("orders: rsi not ready to trade")
	}

	if (stoch.K < 85 || stoch.K > 15) &&
		(stoch.D < 85 || stoch.D > 15) {
		return nil, errors.New("orders: K and D not ready to trade")
	}

	var (
		sideType     futures.SideType
		closeSide    futures.SideType
		positionSide futures.PositionSideType
	)

	if stoch.RSI >= 80 && stoch.K >= 85 && stoch.D >= 85 {
		sideType = futures.SideTypeSell
		closeSide = futures.SideTypeBuy
		positionSide = futures.PositionSideTypeShort
	} else if stoch.RSI <= 15 && stoch.K <= 15 && stoch.D <= 15 {
		sideType = futures.SideTypeBuy
		closeSide = futures.SideTypeSell
		positionSide = futures.PositionSideTypeLong
	}

	if len(candles) < 2 {
		return nil, errors.New("orders: len of candles not enough")
	}

	const (
		gain = 70.0
		loss = 50.0
	)

	var (
		entryPrice float64
	)

	switch positionSide {
	case futures.PositionSideTypeShort:
		entryPrice = helpers.AddFloat(candles[0].High, candles[1].High, currentPrice) / 3.0

		current := helpers.StringToFloat(currentPrice)
		if entryPrice < current {
			entryPrice = current * 0.01
		}

	case futures.PositionSideTypeLong:
		entryPrice = helpers.AddFloat(candles[0].Low, candles[1].Low, currentPrice) / 3.0

		current := helpers.StringToFloat(currentPrice)
		if entryPrice > current {
			entryPrice = current * 0.99
		}
	}

	orders := []*models.Order{
		// DCA the first three
		{
			Symbol:           symbol,
			Side:             sideType,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.Div("10", currentPrice),
			Price:            fmt.Sprint(entryPrice),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		{
			Symbol:           symbol,
			Side:             sideType,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.Div("20", currentPrice),
			Price:            fmt.Sprint(entryPrice * 1.03),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		{
			Symbol:           symbol,
			Side:             sideType,
			PositionSide:     positionSide,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         helpers.Div("30", currentPrice),
			Price:            fmt.Sprint(entryPrice * 1.03 * 1.03),
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
			Quantity:         fmt.Sprint(resolveQuantity(entryPrice*loss, 60)),
			Price:            fmt.Sprint(entryPrice * gain),
			StopPrice:        fmt.Sprint(entryPrice * gain),
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
			Quantity:         fmt.Sprint(resolveQuantity(entryPrice*loss, 60)),
			StopPrice:        fmt.Sprint(entryPrice * loss),
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
	}
	return orders, nil
}

func resolveQuantity(price float64, totalAmount float64) float64 {
	return totalAmount / price
}
