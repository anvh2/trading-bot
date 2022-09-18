package crawler

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
)

var (
	blacklist = map[string]bool{}
)

type Crawler struct {
	logger  *logger.Logger
	binance *futures.Client
	config  *models.ExchangeConfig
	market  *cache.Market
	quit    chan struct{}
}

func New(logger *logger.Logger, config *models.ExchangeConfig, market *cache.Market) *Crawler {
	client := futures.NewClient(config.PublicKey, config.SecretKey)

	return &Crawler{
		logger:  logger,
		binance: client,
		config:  config,
		market:  market,
		quit:    make(chan struct{}),
	}
}

func (c *Crawler) Start() {
	c.WarmUpSymbols()

	go func() {
		c.WarmUpCache()
		c.CrawlData()
		c.logger.Info("[Crawler] warmup success")
	}()
}

func (c *Crawler) Stop() {
	close(c.quit)
}
