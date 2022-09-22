package analyze

import (
	"context"

	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
)

func (s *Server) NotifyPolling(idx int32) {
	for _, symbol := range s.market.Symbols() {
		chart, err := s.market.Chart(symbol)
		if err != nil {
			break
		}

		message := &models.Chart{
			Symbol:     symbol,
			Candles:    make(map[string][]*models.Candlestick),
			UpdateTime: chart.GetUpdateTime(),
		}

		for _, interval := range viper.GetStringSlice("market.intervals") {
			candles, err := chart.Candles(interval)
			if err != nil {
				break
			}

			candleData := candles.Sorted()
			candlesticks := make([]*models.Candlestick, len(candleData))

			for idx, candle := range candleData {
				result, ok := candle.(*models.Candlestick)
				if ok {
					candlesticks[idx] = result
				}
			}

			if len(candlesticks) > 0 {
				message.Candles[interval] = candlesticks
			}
		}

		s.analyze.SendJob(context.Background(), message)
	}
}
