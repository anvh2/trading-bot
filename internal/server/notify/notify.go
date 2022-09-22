package notify

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/anvh2/trading-bot/internal/indicator"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
)

func (s *Server) ProcessAnalyzeCommand(ctx context.Context, args []string) (interface{}, error) {
	if len(args) == 0 {
		return "Empty Symbol", nil
	}

	symbol := args[0]

	oscillator := &models.Oscillator{
		Symbol: symbol,
		Stoch:  make(map[string]*models.Stoch),
	}

	chart, err := s.market.Chart(symbol)
	if err != nil {
		return "Unknown Symbol", nil
	}

	for _, interval := range viper.GetStringSlice("market.intervals") {
		candles, err := chart.Candles(interval)
		if err != nil {
			continue
		}

		candleData := candles.Sorted()

		low := make([]float64, len(candleData))
		high := make([]float64, len(candleData))
		close := make([]float64, len(candleData))

		for idx, data := range candleData {
			candle, ok := data.(*models.Candlestick)
			if ok {
				l, _ := strconv.ParseFloat(candle.Low, 64)
				low[idx] = l

				h, _ := strconv.ParseFloat(candle.High, 64)
				high[idx] = h

				c, _ := strconv.ParseFloat(candle.Close, 64)
				close[idx] = c
			}
		}

		_, rsi := indicator.RSIPeriod(14, close)
		k, d, _ := indicator.KDJ(9, 3, 3, high, low, close)

		stoch := &models.Stoch{
			RSI: rsi[len(rsi)-1],
			K:   k[len(k)-1],
			D:   d[len(d)-1],
		}

		oscillator.Stoch[interval] = stoch
	}

	msg := fmt.Sprintf("%s\n", symbol)

	for _, interval := range viper.GetStringSlice("market.intervals") {
		stoch, ok := oscillator.Stoch[interval]
		if !ok {
			return "Empty Data", nil
		}

		msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.K, stoch.D)
	}

	return msg, nil
}
