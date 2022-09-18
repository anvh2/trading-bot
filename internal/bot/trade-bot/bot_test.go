package tradebot

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
)

func TestPosition(t *testing.T) {
	binance := futures.NewClient("tshhh50wl5HeGOkDXuA4soO81AWyX3AztDb9KoedzZuQ1CSpVidXllJAJzPhXGUB", "KGzctvmH5tsAm4GMTKxbVMwPFybnqIgWBH2rtVgalwyJpM1H2Qax7hyvnYH5i8hy")

	positions, err := binance.NewGetPositionRiskService().Symbol("CHZUSDT").Do(context.Background())
	if err != nil {
		fmt.Println("position error", err)
		return
	}

	pb, _ := json.Marshal(positions)
	fmt.Println(string(pb))

	orders, err := binance.NewListOpenOrdersService().Symbol("BELUSDT").Do(context.Background())
	if err != nil {
		fmt.Println("order error", err)
	}

	ob, _ := json.Marshal(orders)
	fmt.Println(string(ob))

	for _, order := range orders {
		fmt.Println(order.Symbol)
	}
}
