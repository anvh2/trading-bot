package crawler

import (
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/service/binance"
)

var (
	blacklist = map[string]bool{}
)

type Crawler struct {
	logger   *logger.Logger
	binance  *binance.Binance
	market   cache.Market
	exchange cache.Exchange
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

	return ready
}

func (c *Crawler) Stop() {
	close(c.quit)
}
