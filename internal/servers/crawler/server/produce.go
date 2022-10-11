package server

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Server) produce() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("[Produce] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case chart := <-s.message:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				s.publisher.Publish(ctx, "trading.channel.analyze", chart.String())

			case <-ticker.C:
				for _, symbol := range s.exchange.Symbols() {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					chart, err := s.market.Chart(symbol)
					if err != nil {
						continue
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

					s.publisher.Publish(ctx, "trading.channel.analyze", message.String())
				}

			case <-s.quit:
				return
			}
		}
	}()
}
