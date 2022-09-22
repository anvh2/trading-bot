package crawler

import (
	"context"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"golang.org/x/time/rate"
)

func (c *Crawler) warmUpSymbols() error {
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

func (c *Crawler) warmUpCache() error {
	wg := &sync.WaitGroup{}

	limit := rate.NewLimiter(
		rate.Every(viper.GetDuration("binance.rate_limit.duration")),
		viper.GetInt("binance.rate_limit.requests"),
	)

	for _, interval := range viper.GetStringSlice("market.intervals") {
		wg.Add(1)

		go func(interval string) {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("[Crawler][WarmUpCache] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			for _, symbol := range c.market.Symbols() {
				limit.Wait(ctx)

				resp, err := c.futures.ListCandlesticks(ctx, symbol, interval, viper.GetInt("chart.candles.limit"))
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
			}
		}(interval)
	}

	wg.Wait()
	return nil
}
