package crawler

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/cache/exchange"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (c *Crawler) warmUpSymbols() error {
	resp, err := c.binance.GetExchangeInfo(context.Background())
	if err != nil {
		c.logger.Error("[Crawler][WarmUpSymbols] failed to get exchnage info", zap.Error(err))
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

	c.exchange.Set(selected)
	c.logger.Info("[Crawler][WarmUpSymbols] warm up symbols success", zap.Int("total", len(selected)))
	return nil
}

func (c *Crawler) warmUpCache() error {
	wg := &sync.WaitGroup{}

	for _, interval := range viper.GetStringSlice("market.intervals") {
		wg.Add(1)

		go func(interval string) {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("[Crawler][WarmUpCache] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			defer wg.Done()

			for _, symbol := range c.exchange.Symbols() {
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()

				resp, err := c.binance.ListCandlesticks(ctx, symbol, interval, viper.GetInt("chart.candles.limit"))
				if err != nil {
					c.logger.Error("[Crawler][WarmUpCache] failed to get klines data", zap.String("symbol", symbol), zap.String("interval", interval), zap.Error(err))
					c.retry <- &Symbol{Symbol: symbol, Interval: interval}
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

					c.market.UpdateChart(symbol).CreateCandle(interval, candle)
				}

				c.logger.Info("[Crawler][WarmUpCache] success", zap.String("symbol", symbol), zap.String("interval", interval), zap.Int("total", len(resp)))

				time.Sleep(500 * time.Millisecond)
			}
		}(interval)
	}

	wg.Wait()
	return nil
}
