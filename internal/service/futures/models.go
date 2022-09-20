package futures

type Error struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

type CreateOrderResp struct {
	*Error
	OrderId       int    `json:"orderId,omitempty"`
	Symbol        string `json:"symbol,omitempty"`
	Status        string `json:"status,omitempty"`
	ClientOrderId string `json:"clientOrderId,omitempty"`
	Price         string `json:"price,omitempty"`
	AvgPrice      string `json:"avgPrice,omitempty"`
	OrigQty       string `json:"origQty,omitempty"`
	ExecutedQty   string `json:"executedQty,omitempty"`
	CumQty        string `json:"cumQty,omitempty"`
	CumQuote      string `json:"cumQuote,omitempty"`
	TimeInForce   string `json:"timeInForce,omitempty"`
	Type          string `json:"type,omitempty"`
	ReduceOnly    bool   `json:"reduceOnly,omitempty"`
	ClosePosition bool   `json:"closePosition,omitempty"`
	Side          string `json:"side,omitempty"`
	PositionSide  string `json:"positionSide,omitempty"`
	StopPrice     string `json:"stopPrice,omitempty"`
	WorkingType   string `json:"workingType,omitempty"`
	PriceProtect  bool   `json:"priceProtect,omitempty"`
	OrigType      string `json:"origType,omitempty"`
	UpdateTime    int64  `json:"updateTime,omitempty"`
}
