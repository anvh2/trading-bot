package server

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Server) retry() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("[Retry] failed to retry", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		var counter int32 = 0

		for {
			select {
			case symbol := <-s.retryCh:
				delay(&counter)

				resp, err := s.binance.ListCandlesticks(context.Background(), symbol.Symbol, symbol.Interval, viper.GetInt("chart.candles.limit"))
				if err != nil {
					s.logger.Error("[Retry] failed to get klines data", zap.String("symbol", symbol.Symbol), zap.String("interval", symbol.Interval), zap.Error(err))
					s.retryCh <- symbol
					continue
				}

				for _, e := range resp {
					candle := &models.Candlestick{
						OpenTime:  e.OpenTime,
						CloseTime: e.CloseTime,
						Low:       e.Low,
						High:      e.High,
						Close:     e.Close,
					}

					s.market.UpdateChart(symbol.Symbol).CreateCandle(symbol.Interval, candle)
				}

				s.logger.Info("[Retry] success", zap.String("symbol", symbol.Symbol), zap.String("interval", symbol.Interval), zap.Int("total", len(resp)))

				time.Sleep(500 * time.Millisecond)

			case <-s.quit:
				return
			}
		}
	}()
}

func delay(counter *int32) {
	*counter++
	if *counter == 100 {
		*counter = 0
		time.Sleep(30 * time.Minute)
	}
}
