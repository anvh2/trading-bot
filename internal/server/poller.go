package server

import (
	"context"
	"sort"
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
				s.logger.Error("[Polling] failed to process (recovered)", zap.Any("error", r))
			}
		}()

		for {
			select {
			case <-ticker.C:
				for _, symbol := range s.cache.Symbols() {
					message := &models.CandlestickChart{
						Symbol:       symbol,
						Candlesticks: make(map[string][]*models.Candlestick),
					}

					for _, interval := range config.Intervals {
						candleStickData := s.cache.Candlestick(symbol, interval).Range()
						candleSticks := make([]*models.Candlestick, len(candleStickData))

						for idx, candleStick := range candleStickData {
							result, ok := candleStick.(*models.Candlestick)
							if ok {
								candleSticks[idx] = result
							}
						}

						if len(candleSticks) > 0 {
							sort.Slice(candleSticks, func(i, j int) bool {
								return candleSticks[i].OpenTime < candleSticks[j].OpenTime
							})

							message.Candlesticks[interval] = candleSticks
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
