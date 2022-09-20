package crawler

import (
	binance "github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/service/futures"
)

var (
	blacklist = map[string]bool{}
)

type Crawler struct {
	logger  *logger.Logger
	binance *binance.Client
	futures *futures.Futures
	market  *cache.Market
	quit    chan struct{}
}

func New(logger *logger.Logger, market *cache.Market, binance *binance.Client, futures *futures.Futures) *Crawler {
	return &Crawler{
		logger:  logger,
		binance: binance,
		futures: futures,
		market:  market,
		quit:    make(chan struct{}),
	}
}

func (c *Crawler) Start() chan bool {
	ready := make(chan bool)

	c.WarmUpSymbols()

	go func() {
		c.WarmUpCache()
		c.CrawlData()
		ready <- true
	}()

	return ready
}

func (c *Crawler) Stop() {
	close(c.quit)
}
