package cache

import (
	"errors"
	"hash/fnv"
)

var (
	ErrorChartNotFound   = errors.New("chart: not found")
	ErrorCandlesNotFound = errors.New("candles: not found")
)

type Market struct {
	symbols   []string
	cache     []*Chart
	intervals []string
	limit     int32
}

func NewMarket(intervals []string, limit int32) *Market {
	return &Market{
		symbols:   []string{},
		cache:     []*Chart{},
		intervals: intervals,
		limit:     limit,
	}
}

func (c *Market) CacheSymbols(symbols []string) {
	c.symbols = symbols
	c.cache = make([]*Chart, len(symbols))
}

func (c *Market) Symbols() []string {
	return c.symbols
}

func (c *Market) Chart(symbol string) (*Chart, error) {
	idx := c.indexing(symbol)

	if c.cache[idx] == nil {
		return nil, ErrorChartNotFound
	}

	return c.cache[idx], nil
}

func (c *Market) CreateChart(symbol string) *Chart {
	idx := c.indexing(symbol)

	if c.cache[idx] == nil {
		market := new(Chart)
		c.cache[idx] = market.Init(symbol, c.intervals, c.limit)
	}

	return c.cache[idx]
}

func (c *Market) UpdateChart(symbol string) *Chart {
	idx := c.indexing(symbol)

	if c.cache[idx] == nil {
		market := new(Chart)
		c.cache[idx] = market.Init(symbol, c.intervals, c.limit)
	}

	return c.cache[idx]
}

func (c *Market) indexing(symbol string) int {
	return int(hash(symbol)) % len(c.symbols)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
