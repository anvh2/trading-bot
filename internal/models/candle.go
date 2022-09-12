package models

//CandleStick represents a single candle in the graph.
import (
	"encoding/json"
)

//CandleStick represents a single candlestick in a chart.
type Candlestick struct {
	High   string //Represents the highest value obtained during candle period.
	Open   string //Represents the first value of the candle period.
	Close  string //Represents the last value of the candle period.
	Low    string //Represents the lowest value obtained during candle period.
	Volume string //Represents the volume of trades during the candle period.
}

// String returns the string representation of the object.
func (cs Candlestick) String() string {
	b, _ := json.Marshal(cs)
	return string(b)
}

type CandlestickChart struct {
	Symbol       string                    `json:"symbol"`
	Candlesticks map[string][]*Candlestick `json:"candlesticks"`
}
