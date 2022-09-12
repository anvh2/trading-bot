package crawler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

const (
	limit int32 = 500
)

var (
	intervals = []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d"}
)

type Process func(ctx context.Context, message *Message) error

type Message struct {
	Symbol       string                           `json:"symbol"`
	CandleSticks map[string][]*models.CandleStick `json:"candle_sticks"`
}

type Crawler struct {
	logger  *logger.Logger
	binance *binance.Client
	config  *models.ExchangeConfig
	symbols []string
	cache   *cache.Cache
	quit    chan struct{}
	ready   chan bool
	process Process
}

func New(logger *logger.Logger, config *models.ExchangeConfig, process Process) *Crawler {
	client := binance.NewClient(config.PublicKey, config.SecretKey)
	cache := cache.NewCache(&cache.Config{CicularSize: limit})

	crawler := &Crawler{
		logger:  logger,
		binance: client,
		config:  config,
		cache:   cache,
		quit:    make(chan struct{}),
		ready:   make(chan bool),
		process: process,
	}

	crawler.WarmUpSymbols()

	// warmup and connect to websocket in background
	go func() {
		crawler.WarmUpCache()
		crawler.Streaming()
		fmt.Println("Ready to streaming...")
		crawler.ready <- true
	}()

	return crawler
}

func (c *Crawler) Start() {
	ticker := time.NewTicker(time.Millisecond * 1000)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("[Crawler] failed to start, recovered", zap.Any("error", r))
			}
		}()

		<-c.ready

		for {
			select {
			case <-ticker.C:
				for _, symbol := range c.symbols {
					message := &Message{
						Symbol:       symbol,
						CandleSticks: make(map[string][]*models.CandleStick),
					}

					for _, interval := range intervals {
						candleStickData := c.cache.For(symbol, interval).Range()
						candleSticks := make([]*models.CandleStick, len(candleStickData))

						for idx, candleStick := range candleStickData {
							result, ok := candleStick.(*models.CandleStick)
							if ok {
								candleSticks[idx] = result
							}
						}

						message.CandleSticks[interval] = candleSticks
					}

					go c.process(context.Background(), message)
				}

			case <-c.quit:
				return
			}
		}
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
					candle := &models.CandleStick{
						Low:   e.Low,
						High:  e.High,
						Close: e.Close,
					}

					c.cache.For(symbol, interval).Set(candle)
				}

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
	candle := &models.CandleStick{
		Low:   event.Kline.Low,
		High:  event.Kline.High,
		Close: event.Kline.Close,
	}
	c.cache.For(event.Symbol, event.Kline.Interval).Set(candle)
}

func (c *Crawler) handleKlinesStreamError(err error) {
	c.logger.Error("[Crawler][Streaming] failed to recieve stream data", zap.Error(err))
}
