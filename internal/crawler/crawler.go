package crawler

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/internal/service/binance"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	blacklist = map[string]bool{}
)

type Symbol struct {
	Symbol   string
	Interval string
}

type Crawler struct {
	logger   *logger.Logger
	binance  *binance.Binance
	market   cache.Market
	exchange cache.Exchange
	retry    chan *Symbol
	quit     chan struct{}
}

func New(
	logger *logger.Logger,
	market cache.Market,
	exchange cache.Exchange,
	binance *binance.Binance,
) *Crawler {
	return &Crawler{
		logger:   logger,
		binance:  binance,
		market:   market,
		exchange: exchange,
		retry:    make(chan *Symbol, 100),
		quit:     make(chan struct{}),
	}
}

func (c *Crawler) Start() chan bool {
	ready := make(chan bool)

	c.warmUpSymbols()

	go func() {
		c.warmUpCache()
		c.crawlData()
		ready <- true
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Error("[Crawler][Retry] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
			}
		}()

		for {
			select {
			case symbol := <-c.retry:
				resp, err := c.binance.ListCandlesticks(context.Background(), symbol.Symbol, symbol.Interval, viper.GetInt("chart.candles.limit"))
				if err != nil {
					c.logger.Error("[Crawler][Retry] failed to get klines data", zap.String("symbol", symbol.Symbol), zap.String("interval", symbol.Interval), zap.Error(err))
					c.retry <- symbol
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

					c.market.UpdateChart(symbol.Symbol).CreateCandle(symbol.Interval, candle)
				}

				c.logger.Info("[Crawler][Retry] success", zap.String("symbol", symbol.Symbol), zap.String("interval", symbol.Interval), zap.Int("total", len(resp)))

				time.Sleep(500 * time.Millisecond)

			case <-c.quit:
				return
			}
		}
	}()

	return ready
}

func (c *Crawler) Stop() {
	close(c.quit)
}
