package features

import (
	"context"
	"errors"
	"fmt"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-boy/logger"
	"github.com/anvh2/trading-boy/models"
	"go.uber.org/zap"
)

type ClientConfig struct {
	ApiKey    string
	SecretKey string
}

type Exchange struct {
	api    *futures.Client
	logger *logger.Logger
}

func New(logger *logger.Logger, config *ClientConfig) (*Exchange, error) {
	if config == nil {
		return nil, errors.New("invalid client config")
	}

	api := futures.NewClient(config.ApiKey, config.SecretKey)
	return &Exchange{
		api:    api,
		logger: logger,
	}, nil
}

func (e *Exchange) BuyLong(ctx context.Context, market *models.Market, amount int64) error {
	order := e.api.NewCreateOrderService()
	order.Type(futures.OrderTypeLimit).
		Side(futures.SideTypeBuy).
		PositionSide(futures.PositionSideTypeLong).
		Symbol(market.MarketName).
		Quantity(fmt.Sprint(amount))

	resp, err := order.Do(ctx)
	if err != nil {
		e.logger.Error("[Futures][BuyLong] failed to create order", zap.Any("market", market), zap.Int64("amount", amount), zap.Error(err))
		return err
	}

	e.logger.Info("[Futures][BuyLong] create order success", zap.Any("market", market), zap.Int64("amount", amount), zap.Any("resp", resp))
	return nil
}

// func (e *Exchange) BuyShort(ctx context.Context, market *models.Market, amount int64)
