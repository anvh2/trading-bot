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
				for _, symbol := range s.cache.Symbols() {
					market := s.cache.Market(symbol)

					message := &models.CandlesMarket{
						Symbol:       symbol,
						Candlesticks: make(map[string][]*models.Candlestick),
						UpdateTime:   market.Metadata().GetUpdateTime(),
					}

					for _, interval := range config.Intervals {
						candles := market.Candles(interval)
						if candles == nil {
							continue
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
							message.Candlesticks[interval] = candlesticks
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
