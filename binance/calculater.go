package binance

import (
	"context"

	"github.com/anvh2/trading-boy/models"
)

func (bw *BinanceWrapper) CalculateTradingFees(ctx context.Context, market *models.Market, amount float64, limit float64, orderType TradeType) float64 {
	var feePercentage float64
	if orderType == MakerTrade {
		feePercentage = 0.0010
	} else if orderType == TakerTrade {
		feePercentage = 0.0010
	} else {
		panic("Unknown trade type")
	}

	return amount * limit * feePercentage
}

func (bw *BinanceWrapper) CalculateWithdrawFees(ctx context.Context, market *models.Market, amount float64) float64 {
	return 0
}
