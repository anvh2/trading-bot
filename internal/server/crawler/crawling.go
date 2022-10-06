package crawler

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anvh2/trading-bot/internal/cache/exchange"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Server) crawl() error {
	if err := s.crawlExchange(); err != nil {
		return err
	}

	if err := s.crawlMarkets(); err != nil {
		return err
	}

	return nil
}

func (s *Server) crawlExchange() error {
	resp, err := s.binance.GetExchangeInfo(context.Background())
	if err != nil {
		s.logger.Error("[Crawl] failed to get exchange info", zap.Error(err))
		return err
	}

	selected := []*exchange.Symbol{}

	for _, symbol := range resp.Symbols {
		if strings.Contains(symbol.Symbol, "_") {
			continue
		}

		if symbol.MarginAsset == "USDT" {
			if blacklist[symbol.Symbol] {
				continue
			}

			filters := &exchange.Filters{}
			filters.Parse(symbol.Filters)

			selected = append(selected, &exchange.Symbol{
				Symbol:      symbol.Symbol,
				Pair:        symbol.Pair,
				Filters:     filters,
				MarginAsset: symbol.MarginAsset,
				BaseAsset:   symbol.BaseAsset,
			})
		}
	}

	s.exchange.Set(selected)
	s.logger.Info("[Crawl] cache symbols success", zap.Int("total", len(selected)))
	return nil
}

func (s *Server) crawlMarkets() error {
	var (
		wg    = &sync.WaitGroup{}
		total = int32(0)
		start = time.Now()
	)

	for _, interval := range viper.GetStringSlice("market.intervals") {
		wg.Add(1)

		go func(interval string) {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("[Crawl] failed to sync market", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			defer wg.Done()

			for _, symbol := range s.exchange.Symbols() {
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()

				resp, err := s.binance.ListCandlesticks(ctx, symbol, interval, viper.GetInt("chart.candles.limit"))
				if err != nil {
					s.logger.Error("[Crawl] failed to get klines data", zap.String("symbol", symbol), zap.String("interval", interval), zap.Error(err))
					s.retryCh <- &models.Symbol{Symbol: symbol, Interval: interval}
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

					s.market.UpdateChart(symbol).CreateCandle(interval, candle)
				}

				atomic.AddInt32(&total, 1)
				s.logger.Info("[Crawl] cache market success", zap.String("symbol", symbol), zap.String("interval", interval), zap.Int("total", len(resp)))
				time.Sleep(1000 * time.Millisecond)
			}
		}(interval)
	}

	wg.Wait()

	s.logger.Info("[Crawl] success to crawl data", zap.Int32("total", total), zap.Float64("take(s)", time.Since(start).Seconds()))
	return nil
}
