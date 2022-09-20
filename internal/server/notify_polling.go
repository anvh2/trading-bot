package server

import (
	"context"

	"github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/models"
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

		for _, interval := range config.Intervals {
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

		s.notifyWr.SendJob(context.Background(), message)
	}
}
