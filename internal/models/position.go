package models

import "encoding/json"

type OrderStatus string

const (
	OrderStatusNew OrderStatus = "NEW"
)

type OrderSide string

const (
	SideBuy  OrderSide = "BUY"
	SideSell OrderSide = "SELL"
)

type OrderType string

const (
	OrderTypeStopMarket OrderType = "STOP_MARKET"
	OrderTypeTakeProfit OrderType = "TAKE_PROFIT_MARKET"
)

type MarginType string

const (
	MarginTypeCross    MarginType = "cross"
	MarginTypeIsolated MarginType = "isolated"
)

type PositionStatus string

const (
	PositionStatusNew     PositionStatus = "NEW"
	PositionStatusMatched PositionStatus = "MATCHED"
	PositionStatusClosed  PositionStatus = "CLOSED"
)

type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

type Position struct {
	PositionId       string         `json:"position_id"`
	Symbol           string         `json:"symbol"`
	Status           PositionStatus `json:"status"`
	EntryPrice       string         `json:"entry_price"`
	LiquidationPrice string         `json:"liquidation_price"`
	MarkPrice        string         `json:"mark_price"`
	MarginType       MarginType     `json:"margin_type"`
	PositionSide     PositionSide   `json:"position_side"`
	IsolatedWallet   string         `json:"isolated_wallet"`
	UnRealizedProfit string         `json:"un_realized_profit"`
	Leverage         string         `json:"leverage"`
}

func (p *Position) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func (p *Position) Parse(val string) error {
	return json.Unmarshal([]byte(val), p)
}
