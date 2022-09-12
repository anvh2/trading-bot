package models

//CandleStick represents a single candle in the graph.
import (
	"encoding/json"
	"time"
)

//CandleStick represents a single candlestick in a chart.
type CandleStick struct {
	High   string //Represents the highest value obtained during candle period.
	Open   string //Represents the first value of the candle period.
	Close  string //Represents the last value of the candle period.
	Low    string //Represents the lowest value obtained during candle period.
	Volume string //Represents the volume of trades during the candle period.
}

// String returns the string representation of the object.
func (cs CandleStick) String() string {
	b, _ := json.Marshal(cs)
	return string(b)
}

//CandleStickChart represents a chart of a market expresed using Candle Sticks.
type CandleStickChart struct {
	CandlePeriod time.Duration //Represents the candle period (expressed in time.Duration).
	CandleSticks []CandleStick //Represents the last Candle Sticks used for evaluation of current state.
	OrderBook    []Order       //Represents the Book of current trades.
}
