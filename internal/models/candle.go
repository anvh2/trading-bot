package models

//CandleStick represents a single candle in the graph.
import (
	"encoding/json"
)

//CandleStick represents a single candlestick in a chart.
type Candlestick struct {
	OpenTime  int64  `json:"s,omitempty"`
	CloseTime int64  `json:"e,omitempty"`
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

type ChartMetadata struct {
	UpdateTime int64 `json:"update_time"`
}

type CandleChart struct {
	Symbol   string                    `json:"symbol"`
	Candles  map[string][]*Candlestick `json:"candlesticks"` // interval: []candle
	Metadata map[string]*ChartMetadata `json:"metadata"`
}

func (c *CandleChart) String() string {
	if c == nil {
		return ""
	}

	b, _ := json.Marshal(c)
	return string(b)
}

func (c *CandleChart) Len(interval string) int {
	if c == nil {
		return 0
	}

	return len(c.Candles[interval])
}
