package crawler

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

func (c *Crawler) WarmUpSymbols() error {
	resp, err := c.binance.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		c.logger.Error("[Crawler][WarmUpSymbols] failed to get exchnage info", zap.Error(err))
		return err
	}

	selected := []string{}

	for _, symbol := range resp.Symbols {
		if strings.Contains(symbol.Symbol, "USDT") {
			if blacklist[symbol.Symbol] {
				continue
			}
			selected = append(selected, symbol.Symbol)
		}
	}

	c.market.CacheSymbols(selected)
	c.logger.Info("[Crawler][WarmUpSymbols] warm up symbols success", zap.Int("total", len(selected)))
	return nil
}

func (c *Crawler) WarmUpCache() error {
	wg := &sync.WaitGroup{}

	for _, interval := range config.Intervals {
		wg.Add(1)

		go func(interval string) {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("[Crawler][WarmUpCache] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			defer wg.Done()

			for _, symbol := range c.market.Symbols() {
				resp, err := c.futures.ListCandlesticks(context.Background(), symbol, interval, int(config.CandleLimit))
				if err != nil {
					c.logger.Error("[Crawler][WarmUpCache] failed to get klines data", zap.String("symbol", symbol), zap.String("interval", interval), zap.Error(err))
					return
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
				time.Sleep(time.Millisecond * 500) // TODO: temporary rate limit for calling binance api, default allow 1200 per minute
			}
		}(interval)
	}

	wg.Wait()
	return nil
}
