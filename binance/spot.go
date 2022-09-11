package binance

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-boy/models"
)

func (bw *BinanceWrapper) BuyLimit(ctx context.Context, market *models.Market, amount float64, limit float64) (string, error) {
	orderSrv := bw.api.NewCreateOrderService()
	orderNumber, err := orderSrv.
		Type(binance.OrderTypeLimit).
		Side(binance.SideTypeBuy).
		Symbol(market.MarketName).
		Price(fmt.Sprint(limit)).
		Quantity(fmt.Sprint(amount)).
		Do(ctx)
	if err != nil {
		return "", err
	}

	return orderNumber.ClientOrderID, nil
}

func (bw *BinanceWrapper) SellLimit(ctx context.Context, market *models.Market, amount float64, limit float64) (string, error) {
	orderSrv := bw.api.NewCreateOrderService()
	orderNumber, err := orderSrv.
		Type(binance.OrderTypeLimit).
		Side(binance.SideTypeSell).
		Symbol(market.MarketName).
		Price(fmt.Sprint(limit)).
		Quantity(fmt.Sprint(amount)).
		Do(ctx)
	if err != nil {
		return "", err
	}

	return orderNumber.ClientOrderID, nil
}

func (bw *BinanceWrapper) BuyMarket(ctx context.Context, market *models.Market, amount float64) (string, error) {
	orderSrv := bw.api.NewCreateOrderService()
	orderNumber, err := orderSrv.
		Type(binance.OrderTypeMarket).
		Side(binance.SideTypeBuy).
		Symbol(market.MarketName).
		Quantity(fmt.Sprint(amount)).
		Do(ctx)
	if err != nil {
		return "", err
	}

	return orderNumber.ClientOrderID, nil
}

func (bw *BinanceWrapper) SellMarket(ctx context.Context, market *models.Market, amount float64) (string, error) {
	orderSrv := bw.api.NewCreateOrderService()
	orderNumber, err := orderSrv.
		Type(binance.OrderTypeMarket).
		Side(binance.SideTypeSell).
		Symbol(market.MarketName).
		Quantity(fmt.Sprint(amount)).
		Do(ctx)
	if err != nil {
		return "", err
	}

	return orderNumber.ClientOrderID, nil
}
