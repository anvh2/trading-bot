package models

import (
	"encoding/json"

	"github.com/adshao/go-binance/v2/futures"
)

type Order struct {
	Symbol           string
	Side             futures.SideType
	PositionSide     futures.PositionSideType
	OrderType        futures.OrderType
	TimeInForce      futures.TimeInForceType
	Quantity         string
	ReduceOnly       bool
	Price            string
	NewClientOrderId string // callback id
	StopPrice        string
	WorkingType      futures.WorkingType
	ActivationPrice  string
	CallbackRate     string
	PriceProtect     bool
	NewOrderRespType futures.NewOrderRespType
	ClosePosition    bool
}

func (o *Order) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *Order) Parse(val string) error {
	return json.Unmarshal([]byte(val), o)
}
