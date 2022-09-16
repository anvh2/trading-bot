package crawler

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

var (
	blacklist = map[string]bool{"LUNCUSDT": true, "VENUSDT": true}
)

type Crawler struct {
	logger  *logger.Logger
	binance *binance.Client
	config  *models.ExchangeConfig
	cache   *cache.Cache
	quit    chan struct{}
}

func New(logger *logger.Logger, config *models.ExchangeConfig, cache *cache.Cache) *Crawler {
	client := binance.NewClient(config.PublicKey, config.SecretKey)

	return &Crawler{
		logger:  logger,
		binance: client,
		config:  config,
		cache:   cache,
		quit:    make(chan struct{}),
	}
}

func (c *Crawler) Start() {
	c.WarmUpSymbols()

	go func() {
		c.WarmUpCache()
		c.Streaming()
		fmt.Println("Streaming data...")
	}()
}

func (c *Crawler) Stop() {
	close(c.quit)
}

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

	c.cache.SetSymbols(selected)
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

			for _, symbol := range c.cache.Symbols() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err := c.binance.NewKlinesService().Symbol(symbol).Interval(interval).Limit(int(config.CandleSize)).Do(ctx)
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

					c.cache.Market(symbol).CreateCandle(interval, candle)
				}

				c.logger.Info("[Crawler][WarmUpCache] success", zap.String("symbol", symbol), zap.String("interval", interval), zap.Int("total", len(resp)))
				time.Sleep(time.Millisecond * 500) // TODO: temporary rate limit for calling binance api, default allow 1200 per minute
			}
		}(interval)
	}

	wg.Wait()
	return nil
}

func (c *Crawler) Streaming() error {
	for _, interval := range config.Intervals {
		pair := make(map[string]string, len(c.cache.Symbols()))
		for _, symbol := range c.cache.Symbols() {
			pair[symbol] = interval
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("[Crawler][Streaming] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			done, _, err := binance.WsCombinedKlineServe(pair, c.handleKlinesStreamData, c.handleKlinesStreamError)
			if err != nil {
				c.logger.Error("[Crawler][Streaming] failed to connect to klines stream data", zap.Error(err))
				return
			}

			<-done
		}()

		time.Sleep(5 * time.Second)
	}

	return nil
}

func (c *Crawler) handleKlinesStreamData(event *binance.WsKlineEvent) {
	market := c.cache.Market(event.Symbol)
	candles := market.Candles(event.Kline.Interval)
	if candles == nil {
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

	// update last candle
	if lastCandle.OpenTime == event.Kline.StartTime &&
		lastCandle.CloseTime == event.Kline.EndTime {

		lastCandle.Close = event.Kline.Close

		candles.Update(idx, lastCandle)
		market.UpdateMeta()
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

	market.CreateCandle(event.Kline.Interval, candle)
}

func (c *Crawler) handleKlinesStreamError(err error) {
	c.logger.Error("[Crawler][Streaming] failed to recieve stream data", zap.Error(err))
}
