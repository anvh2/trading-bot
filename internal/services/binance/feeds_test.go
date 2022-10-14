package binance

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExchangeInfo(t *testing.T) {
	resp, err := test_binanceInst.GetExchangeInfo(context.Background())
	assert.Nil(t, err)

	fmt.Println(resp)
}

func TestGetCurrentPrice(t *testing.T) {
	resp, err := test_binanceInst.GetCurrentPrice(context.Background(), "BTCUSDT")
	assert.Nil(t, err)
	fmt.Println(resp)
}

func TestListCandlesticks(t *testing.T) {
	resp, err := test_binanceInst.ListCandlesticks(context.TODO(), "BTCUSDT", "1h", 10)
	assert.Nil(t, err)

	fmt.Println(resp)
}
