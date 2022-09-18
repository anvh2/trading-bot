package server

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/models"

	"go.uber.org/zap"
)

func (s *Server) polling() {
	ticker := time.NewTicker(time.Millisecond * 1000)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("[Polling] failed to process (recovered)", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		for {
			select {
			case <-ticker.C:
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

					s.worker.SendJob(context.Background(), message)
				}

			case <-s.quitPolling:
				return
			}
		}
	}()

}
