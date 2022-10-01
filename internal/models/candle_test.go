package models

import (
	"fmt"
	"testing"
)

func TestMarshalChart(t *testing.T) {
	chart := &Chart{
		Symbol: "BTCUSDT",
		Candles: map[string][]*Candlestick{
			"1h": {
				{
					Low:  "10",
					High: "20",
				},
			},
		},
	}
	fmt.Println(chart.String())
}
