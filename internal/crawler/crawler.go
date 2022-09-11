package crawler

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-boy/internal/logger"
	"github.com/anvh2/trading-boy/internal/models"
	"go.uber.org/zap"
)

const (
	limit int32 = 500
)

var (
	intervals = []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d"}
)

type Process func(ctx context.Context, message []interface{}) error

type Crawler struct {
	logger  *logger.Logger
	binance *binance.Client
	config  *models.ExchangeConfig
	symbols []string
	cache   *Cache
	quitCh  chan struct{}
	process Process
}

func New(logger *logger.Logger, config *models.ExchangeConfig, process Process) *Crawler {
	client := binance.NewClient(config.PublicKey, config.SecretKey)

	crawler := &Crawler{
		logger:  logger,
		binance: client,
		config:  config,
		cache:   NewCache(),
		quitCh:  make(chan struct{}),
		process: process,
	}

	crawler.WarmUpSymbols()

	go func() {
		crawler.WarmUpCache()
		crawler.Streaming()
	}()

	return crawler
}

func (c *Crawler) Start() {
	ticker := time.NewTicker(time.Millisecond * 10000)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("[Crawler] failed to start, recovered", zap.Any("error", r))
			}
		}()

		for {
			select {
			case <-ticker.C:
				data := c.cache.For("BTCUSDT", "1m").Range()
				c.process(context.Background(), data)

			case <-c.quitCh:
				return
			}
		}
	}()
}

func (c *Crawler) Stop() {
	close(c.quitCh)
}

func (c *Crawler) WarmUpSymbols() error {
	resp, err := c.binance.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		c.logger.Error("[Crawler][WarmUpSymbols] failed to get exchnage info", zap.Error(err))
		return err
	}

	for _, symbol := range resp.Symbols {
		if strings.Contains(symbol.Symbol, "USDT") {
			c.symbols = append(c.symbols, symbol.Symbol)
		}
	}

	c.logger.Info("[Crawler][WarmUpSymbols] warm up symbols success", zap.Int("total", len(c.symbols)))
	return nil
}

func (c *Crawler) WarmUpCache() error {
	wg := &sync.WaitGroup{}

	for _, interval := range intervals {
		wg.Add(1)

		go func(interval string) {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("[Crawler][WarmUpCache] failed to start, recovered", zap.Any("error", r))
				}
			}()

			defer wg.Done()

			for _, symbol := range c.symbols {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err := c.binance.NewKlinesService().Symbol(symbol).Interval(interval).Limit(int(limit)).Do(ctx)
				if err != nil {
					c.logger.Error("[Crawler][WarmUpCache] failed to get klines data", zap.String("symbol", symbol), zap.String("interval", interval), zap.Error(err))
					return
				}

				for _, e := range resp {
					c.cache.For(symbol, interval).Set(e.Close)
				}

				c.logger.Info("[WarmUpCache] warm up success", zap.String("symbol", symbol), zap.String("interval", interval))
				time.Sleep(time.Millisecond * 500) // TODO: temporary rate limit for calling binance api, default allow 1200 per minute
			}
		}(interval)
	}

	wg.Wait()
	return nil
}

func (c *Crawler) Streaming() error {
	for _, interval := range intervals {
		pair := make(map[string]string, len(c.symbols))
		for _, symbol := range c.symbols {
			pair[symbol] = interval
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("[Crawler][Streaming] failed to start, recovered", zap.Any("error", r))
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
	c.cache.For(event.Symbol, event.Kline.Interval).Set(event.Kline.Close)
}

func (c *Crawler) handleKlinesStreamError(err error) {
	c.logger.Error("[Crawler][Streaming] failed to recieve stream data", zap.Error(err))
}
