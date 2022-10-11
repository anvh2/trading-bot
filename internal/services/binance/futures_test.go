package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
)

func TestListPositionRisk(t *testing.T) {
	viper.SetConfigFile("../../../config.dev.toml")
	viper.ReadInConfig()

	f := New(logger.NewDev())

	resp, err := f.ListPositionRisk(context.Background(), "XRPUSDT")
	fmt.Println(resp, err)
}

func TestCreateOrders(t *testing.T) {
	viper.SetConfigFile("../../../config.dev.toml")
	viper.ReadInConfig()

	f := New(logger.NewDev())

	resp, err := f.CreateOrders(context.Background(), []*models.Order{
		{
			Symbol:           "BNXUSDT",
			Side:             futures.SideTypeSell,
			PositionSide:     futures.PositionSideTypeShort,
			OrderType:        futures.OrderTypeLimit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         "0.1",
			Price:            "170",
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeRESULT,
		},
		// {
		// 	Symbol:           "BNXUSDT",
		// 	Side:             futures.SideTypeSell,
		// 	PositionSide:     futures.PositionSideTypeShort,
		// 	OrderType:        futures.OrderTypeTakeProfit,
		// 	TimeInForce:      futures.TimeInForceTypeGTC,
		// 	Quantity:         "0.1",
		// 	Price:            "170",
		// 	StopPrice:        "170",
		// 	WorkingType:      futures.WorkingTypeMarkPrice,
		// 	NewOrderRespType: futures.NewOrderRespTypeACK,
		// },
		// {
		// 	Symbol:           "BNXUSDT",
		// 	Side:             futures.SideTypeSell,
		// 	PositionSide:     futures.PositionSideTypeShort,
		// 	OrderType:        futures.OrderTypeStopMarket,
		// 	TimeInForce:      futures.TimeInForceTypeGTC,
		// 	Quantity:         "0.1",
		// 	StopPrice:        "120",
		// 	WorkingType:      futures.WorkingTypeMarkPrice,
		// 	NewOrderRespType: futures.NewOrderRespTypeACK,
		// },
	})
	b, _ := json.Marshal(resp)
	fmt.Println(string(b), err)
}
