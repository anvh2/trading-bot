package binance

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

func (bw *BinanceWrapper) GetBalance(ctx context.Context, symbol string) (*decimal.Decimal, error) {
	binanceAccount, err := bw.api.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, err
	}

	for _, binanceBalance := range binanceAccount.Balances {
		if binanceBalance.Asset == symbol {
			ret, err := decimal.NewFromString(binanceBalance.Free)
			if err != nil {
				return nil, err
			}
			return &ret, nil
		}
	}

	return nil, errors.New("symbol not found")
}
