package trader

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
)

func TestPosition(t *testing.T) {
	binance := futures.NewClient("tshhh50wl5HeGOkDXuA4soO81AWyX3AztDb9KoedzZuQ1CSpVidXllJAJzPhXGUB", "KGzctvmH5tsAm4GMTKxbVMwPFybnqIgWBH2rtVgalwyJpM1H2Qax7hyvnYH5i8hy")

	positions, err := binance.NewGetPositionRiskService().Symbol("XLMUSDT").Do(context.Background())
	if err != nil {
		fmt.Println("position error", err)
		return
	}

	pb, _ := json.Marshal(positions)
	fmt.Println(string(pb))

	orders, err := binance.NewListOpenOrdersService().Symbol("XLMUSDT").Do(context.Background())
	if err != nil {
		fmt.Println("order error", err)
	}

	ob, _ := json.Marshal(orders)
	fmt.Println(string(ob))

	// currentPrice, err := binance.NewListPricesService().Symbol("BTCUSDT").Do(context.Background())
	// if err != nil {
	// 	fmt.Println("price error", err)
	// }

	// fmt.Println(currentPrice[0])

	// candles, err := binance.NewKlinesService().Symbol("BTCUSDT").Interval("1h").Limit(1).Do(context.Background())
	// if err != nil {
	// 	fmt.Println("candles error", err)
	// }

	// b, _ := json.Marshal(candles[0])
	// fmt.Println(string(b), len(candles))
}

func TestCreateOrder(t *testing.T) {
	binance := futures.NewClient("tshhh50wl5HeGOkDXuA4soO81AWyX3AztDb9KoedzZuQ1CSpVidXllJAJzPhXGUB", "KGzctvmH5tsAm4GMTKxbVMwPFybnqIgWBH2rtVgalwyJpM1H2Qax7hyvnYH5i8hy")

	order := binance.NewCreateOrderService().
		Symbol("BNXUSDT").
		Side(futures.SideTypeSell).
		PositionSide(futures.PositionSideTypeShort).
		Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity("0.1").
		Price("170").
		WorkingType(futures.WorkingTypeMarkPrice).
		NewOrderResponseType(futures.NewOrderRespTypeRESULT)

	resp, err := binance.NewCreateBatchOrdersService().OrderList([]*futures.CreateOrderService{order}).Do(context.Background())
	fmt.Println(resp, err)
}

func TestTakeProfit(t *testing.T) {
	binance := futures.NewClient("tshhh50wl5HeGOkDXuA4soO81AWyX3AztDb9KoedzZuQ1CSpVidXllJAJzPhXGUB", "KGzctvmH5tsAm4GMTKxbVMwPFybnqIgWBH2rtVgalwyJpM1H2Qax7hyvnYH5i8hy")

	order := &futures.CreateOrderService{}

	reduceOnly := true

	order.Symbol("BNXUSDT").
		Side(futures.SideTypeSell).
		PositionSide(futures.PositionSideTypeShort).
		Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity("0.1").
		ReduceOnly(reduceOnly).
		Price("170").
		StopPrice("170").
		WorkingType(futures.WorkingTypeMarkPrice).
		NewOrderResponseType(futures.NewOrderRespTypeACK)

	resp, err := binance.NewCreateBatchOrdersService().OrderList([]*futures.CreateOrderService{order}).Do(context.Background())
	fmt.Println(resp, err)
}
