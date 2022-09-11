package crawler

import (
	"fmt"
	"sync"

	"github.com/anvh2/trading-boy/internal/cache/circular"
)

type Cache struct {
	mutex *sync.Mutex
	cache map[string]*circular.Cache // map[symbol-interval]close_prices
	key   func(val1, val2 string) string
}

func NewCache() *Cache {
	return &Cache{
		mutex: &sync.Mutex{},
		cache: make(map[string]*circular.Cache),
		key: func(symbol, interval string) string {
			return fmt.Sprintf("%s-%s", symbol, interval)
		},
	}
}

func (c *Cache) For(symbol, interval string) *circular.Cache {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache[c.key(symbol, interval)] == nil {
		c.cache[c.key(symbol, interval)] = circular.New(limit)
	}

	return c.cache[c.key(symbol, interval)]
}
