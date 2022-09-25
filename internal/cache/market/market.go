package market

import (
	"sync"

	"github.com/anvh2/trading-bot/internal/cache/circular"
	"github.com/anvh2/trading-bot/internal/cache/errors"
)

type Market struct {
	mutex *sync.Mutex
	cache map[string]*Chart // map[symbol]chart
	limit int32
}

func NewMarket(limit int32) *Market {
	return &Market{
		mutex: &sync.Mutex{},
		cache: make(map[string]*Chart),
		limit: limit,
	}
}

func (c *Market) Chart(symbol string) (*Chart, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		return nil, errors.ErrorChartNotFound
	}

	return c.cache[symbol], nil
}

func (c *Market) CreateChart(symbol string) *Chart {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(Chart)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	return c.cache[symbol]
}

func (c *Market) UpdateChart(symbol string) *Chart {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(Chart)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	return c.cache[symbol]
}

func (c *Market) Candles(symbol, interval string) *circular.Cache {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(Chart)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	candles := c.cache[symbol].cache[interval]
	if candles == nil {
		candles = circular.New(c.limit)
	}

	return candles
}
