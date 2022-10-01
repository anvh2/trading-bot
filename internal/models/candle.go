package models

//CandleStick represents a single candle in the graph.
import (
	"encoding/json"
)

//CandleStick represents a single candlestick in a chart.
type Candlestick struct {
	OpenTime  int64  `json:"ot,omitempty"`
	CloseTime int64  `json:"ct,omitempty"`
	High      string `json:"h,omitempty"`
	Open      string `json:"o,omitempty"`
	Close     string `json:"c,omitempty"`
	Low       string `json:"l,omitempty"`
	Volume    string `json:"v,omitempty"`
}

// String returns the string representation of the object.
func (cs *Candlestick) String() string {
	b, _ := json.Marshal(cs)
	return string(b)
}

type Chart struct {
	Symbol     string                    `json:"symbol"`
	Candles    map[string][]*Candlestick `json:"candlesticks"`
	UpdateTime int64                     `json:"update_time"`
}

func (chart *Chart) String() string {
	b, _ := json.Marshal(chart)
	return string(b)
}
