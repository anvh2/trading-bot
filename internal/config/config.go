package config

var (
	Intervals = []string{"5m", "15m", "30m", "1h", "4h", "1d"}
)

const (
	CandleLimit int32 = 1000

	TelegramChatId int64 = -653827904 // Trading Recommendation
	// TelegramChatId int64 = 1630847448 // @anvh21
)

//SPDR S&P 500 Trust (SPY) shows different Stochastics footprints, depending on variables.
// Cycle turns occur when the fast line crosses the slow line after reaching the overbought
// or oversold level. The responsive 5,3,3 setting flips buy and sell cycles frequently,
// often without the lines reaching overbought or oversold levels. The mid-range 21,7,7
// setting looks back at a longer period but keeps smoothing at relatively low levels,
// yielding wider swings that generate fewer buy and sell signals. The long-term 21,14,14
// setting takes a giant step back, signaling cycle turns rarely and only near key market turning points.
// Refer: https://www.investopedia.com/articles/active-trading/021915/pick-right-settings-your-stochastic-oscillator.asp
type StochType int32 // trade style

const (
	StochShortTerm     = 1
	StochMediumTerm    = 2
	StochLongTerm      = 3
	StochSuperLongTerm = 4
)

const (
	RSIPeriod int = 14
)

type StochSetting struct {
	FastKPeriod int
	SlowKPeriod int
	SlowDPeriod int
}

var (
	StochSettings = map[StochType]*StochSetting{
		StochShortTerm:     {9, 3, 3},
		StochMediumTerm:    {12, 3, 3},
		StochLongTerm:      {12, 3, 15},
		StochSuperLongTerm: {21, 7, 7},
	}
)
