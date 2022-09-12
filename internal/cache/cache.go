package cache

import (
	"fmt"
	"sync"

	"github.com/anvh2/trading-bot/internal/cache/circular"
)

type Config struct {
	CicularSize int32
}

type Cache struct {
	mutex   *sync.Mutex
	symbols []string
	candles map[string]*circular.Cache // map[symbol-interval]close_prices
	config  *Config
	key     func(val1, val2 string) string
}

func NewCache(config *Config) *Cache {
	if config == nil {
		config = &Config{
			CicularSize: 500,
		}
	}

	return &Cache{
		mutex:   &sync.Mutex{},
		candles: make(map[string]*circular.Cache),
		config:  config,
		key: func(symbol, interval string) string {
			return fmt.Sprintf("%s-%s", symbol, interval)
		},
	}
}

func (c *Cache) SetSymbols(symbols []string) {
	c.symbols = symbols
}

func (c *Cache) Symbols() []string {
	return c.symbols
}

func (c *Cache) Candlestick(symbol, interval string) *circular.Cache {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.candles[c.key(symbol, interval)] == nil {
		c.candles[c.key(symbol, interval)] = circular.New(c.config.CicularSize)
	}

	return c.candles[c.key(symbol, interval)]
}
