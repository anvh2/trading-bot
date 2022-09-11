package crawler

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-boy/internal/cache/circular"
	"github.com/anvh2/trading-boy/internal/logger"
	"github.com/anvh2/trading-boy/internal/models"
	"github.com/markcheno/go-talib"
	"go.uber.org/zap"
)

const (
	limit int32 = 500
)

var (
	intervals = []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d"}
)

type Crawler struct {
	logger  *logger.Logger
	binance *binance.Client
	config  *models.ExchangeConfig
	symbols []string
	cache   map[string]*circular.Cache // map[symbol-interval]close_prices
	quitCh  chan struct{}
}

func cacheKey(symbol, interval string) string {
	return fmt.Sprintf("%s-%s", symbol, interval)
}

func New(logger *logger.Logger, config *models.ExchangeConfig, symbols []string) *Crawler {
	client := binance.NewClient(config.PublicKey, config.SecretKey)

	crawler := &Crawler{
		logger:  logger,
		binance: client,
		config:  config,
		symbols: symbols,
		cache:   make(map[string]*circular.Cache),
		quitCh:  make(chan struct{}),
	}

	crawler.WarmUp()
	go crawler.Streaming()

	return crawler
}

func (c *Crawler) Start() {
	ticker := time.NewTicker(time.Millisecond * 1000)
	current := 0.0
	next := 0.0
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("[Crawler] failed to start, recovered", zap.Any("error", r))
			}
		}()

		for {
			select {
			case <-ticker.C:
				current = next

				data := c.cache["BTCUSDT-1m"].Range()

				inputs := []float64{}
				for _, e := range data {
					input, _ := strconv.ParseFloat(e, 64)
					inputs = append(inputs, input)
				}

				rsi := talib.Rsi(inputs, 14)
				next = rsi[len(rsi)-1]

				if current != next {
					fmt.Println("RSI BTCUSDT-1m:\n current: ", current, ", next: ", next)
				}

			case <-c.quitCh:
				return
			}
		}
	}()
}

func (c *Crawler) WarmUp() error {
	wg := &sync.WaitGroup{}

	for _, interval := range intervals {
		for _, symbol := range c.symbols {
			wg.Add(1)

			go func(symbol, interval string) {
				defer func() {
					if r := recover(); r != nil {
						c.logger.Error("[Crawler][WarmUp] failed to start, recovered", zap.Any("error", r))
					}
				}()

				defer wg.Done()

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				resp, err := c.binance.NewKlinesService().Symbol(symbol).Interval(interval).Limit(int(limit)).Do(ctx)
				if err != nil {
					c.logger.Fatal("[Crawler][WarmUp] failed to get klines data", zap.String("symbol", symbol), zap.String("interval", interval), zap.Error(err))
					return
				}

				key := cacheKey(symbol, interval)
				if c.cache[key] == nil {
					c.cache[key] = circular.New(limit)
				}

				for _, e := range resp {
					c.cache[key].Set(e.Close)
				}

				c.logger.Info("[WarmUp] warm up success", zap.String("symbol", symbol), zap.String("interval", interval))

			}(symbol, interval)
		}
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
	}

	return nil
}

func (c *Crawler) handleKlinesStreamData(event *binance.WsKlineEvent) {
	key := cacheKey(event.Symbol, event.Kline.Interval)
	if c.cache[key] == nil {
		c.cache[key] = circular.New(limit)
	}

	c.cache[key].Set(event.Kline.Close)
	c.logger.Info("[Crawler][Streaming] recieve data from data feed", zap.Any("data", event))
}

func (c *Crawler) handleKlinesStreamError(err error) {
	c.logger.Error("[Crawler][Streaming] failed to recieve stream data", zap.Error(err))
}
