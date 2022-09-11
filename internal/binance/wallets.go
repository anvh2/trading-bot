package binance

import (
	"context"
	"fmt"
)

func (bw *BinanceWrapper) Withdraw(ctx context.Context, destinationAddress string, coinTicker string, amount float64) error {
	withdrawSrv := bw.api.NewCreateWithdrawService()
	_, err := withdrawSrv.
		Address(destinationAddress).
		Coin(coinTicker).
		Amount(fmt.Sprint(amount)).
		Do(ctx)
	if err != nil {
		return err
	}

	return nil
}
