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
	f := New(logger.NewDev(), nil, &Config{ApiKey: "tshhh50wl5HeGOkDXuA4soO81AWyX3AztDb9KoedzZuQ1CSpVidXllJAJzPhXGUB", SecretKey: "KGzctvmH5tsAm4GMTKxbVMwPFybnqIgWBH2rtVgalwyJpM1H2Qax7hyvnYH5i8hy"})

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
