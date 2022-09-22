package futures

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
)

func TestCreateOrders(t *testing.T) {
	f := New(logger.NewDev(), nil, &Config{ApiKey: "24a91b0057cd5df6f3fa4c9e059511670d84951b8b1e4cb3eb725b75b7a855bc", SecretKey: "c6fcd23215be8e792ce8262e5f2c180abdb4aa9ec832d1c54bacad90e0437ae7"})

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
		{
			Symbol:           "BNXUSDT",
			Side:             futures.SideTypeSell,
			PositionSide:     futures.PositionSideTypeShort,
			OrderType:        futures.OrderTypeTakeProfit,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         "0.1",
			Price:            "170",
			StopPrice:        "170",
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeACK,
		},
		{
			Symbol:           "BNXUSDT",
			Side:             futures.SideTypeSell,
			PositionSide:     futures.PositionSideTypeShort,
			OrderType:        futures.OrderTypeStopMarket,
			TimeInForce:      futures.TimeInForceTypeGTC,
			Quantity:         "0.1",
			StopPrice:        "120",
			WorkingType:      futures.WorkingTypeMarkPrice,
			NewOrderRespType: futures.NewOrderRespTypeACK,
		},
	})
	b, _ := json.Marshal(resp)
	fmt.Println(string(b), err)
}
