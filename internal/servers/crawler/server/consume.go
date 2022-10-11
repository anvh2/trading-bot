package server

import (
	"runtime/debug"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/cache/errors"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Server) consume() error {
	for _, interval := range viper.GetStringSlice("market.intervals") {
		pair := make(map[string]string, len(s.exchange.Symbols()))
		for _, symbol := range s.exchange.Symbols() {
			pair[symbol] = interval
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("[Consume] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			done, _, err := futures.WsCombinedKlineServe(pair, s.handleConsumeCandles, s.handleConsumeError)
			if err != nil {
				s.logger.Fatal("[Consume] failed to connect to klines stream data", zap.Error(err))
				return
			}

			<-done
		}()

		time.Sleep(2 * time.Second)
	}

	return nil
}

func (s *Server) handleConsumeCandles(event *futures.WsKlineEvent) {
	_, err := s.exchange.Get(event.Symbol)
	if err == errors.ErrorSymbolNotFound {
		s.logger.Info("[Consume] no need to handle this symbol", zap.String("symbol", event.Symbol))
		return
	}

	chart, err := s.market.Chart(event.Symbol)
	if err == errors.ErrorChartNotFound {
		chart = s.market.CreateChart(event.Symbol)
	}

	candles, err := chart.Candles(event.Kline.Interval)
	if err == errors.ErrorCandlesNotFound {
		return
	}

	last, idx := candles.Last()
	if idx < 0 {
		return
	}

	lastCandle, ok := last.(*models.Candlestick)
	if !ok {
		return
	}

	// update the last candle
	if lastCandle.OpenTime == event.Kline.StartTime &&
		lastCandle.CloseTime == event.Kline.EndTime {

		lastCandle.Close = event.Kline.Close
		lastCandle.High = event.Kline.High
		lastCandle.Low = event.Kline.Low

		chart.UpdateCandle(event.Kline.Interval, idx, lastCandle)
		return
	}

	// create new candle
	candle := &models.Candlestick{
		OpenTime:  event.Kline.StartTime,
		CloseTime: event.Kline.EndTime,
		Low:       event.Kline.Low,
		High:      event.Kline.High,
		Close:     event.Kline.Close,
	}

	chart.CreateCandle(event.Kline.Interval, candle)
}

func (s *Server) handleConsumeError(err error) {
	s.logger.Error("[Consume] failed to recieve data", zap.Error(err))
}
