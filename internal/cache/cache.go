package cache

import (
	"sync"

	"github.com/anvh2/trading-bot/internal/cache/circular"
)

type Cache struct {
	mutex   *sync.Mutex
	symbols []string
	cache   map[string]*Market // map[symbol]market
	limit   int32
}

func NewCache(limit int32) *Cache {
	return &Cache{
		mutex:   &sync.Mutex{},
		symbols: []string{},
		cache:   make(map[string]*Market),
		limit:   limit,
	}
}

func (c *Cache) SetSymbols(symbols []string) {
	c.symbols = symbols
}

func (c *Cache) Symbols() []string {
	return c.symbols
}

func (c *Cache) Market(symbol string) *Market {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(Market)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	return c.cache[symbol]
}

func (c *Cache) Candles(symbol, interval string) *circular.Cache {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[symbol] == nil {
		market := new(Market)
		c.cache[symbol] = market.Init(symbol, c.limit)
	}

	candles := c.cache[symbol].candles[interval]
	if candles == nil {
		candles = circular.New(c.limit)
	}

	return candles
}
