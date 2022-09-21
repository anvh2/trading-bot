package cache

import (
	"time"

	"github.com/anvh2/trading-bot/internal/cache/circular"
	"github.com/anvh2/trading-bot/internal/models"
)

type ChartMeta struct {
	UpdateTime int64
}

type Chart struct {
	symbol    string // key of chart
	intervals []string
	cache     []*circular.Cache // map[interval]candles
	meta      *ChartMeta
	limit     int32 // limit of candles's length
}

func (m *Chart) Init(symbol string, intervals []string, limit int32) *Chart {
	return &Chart{
		symbol:    symbol,
		intervals: intervals,
		cache:     make([]*circular.Cache, len(intervals)),
		meta:      &ChartMeta{},
		limit:     limit,
	}
}

func (m *Chart) Candles(interval string) (*circular.Cache, error) {
	idx := m.indexing(interval)

	if m.cache[idx] == nil {
		return nil, ErrorCandlesNotFound
	}

	return m.cache[idx], nil
}

func (m *Chart) CreateCandle(interval string, candle *models.Candlestick) error {
	idx := m.indexing(interval)

	if m.cache[idx] == nil {
		m.cache[idx] = circular.New(m.limit)
	}

	m.cache[idx].Create(candle)
	m.meta.UpdateTime = time.Now().UnixMilli()

	return nil
}

func (m *Chart) UpdateCandle(interval string, candleId int32, candle *models.Candlestick) error {
	idx := m.indexing(interval)

	if m.cache[idx] == nil {
		return ErrorCandlesNotFound
	}

	m.cache[idx].Update(candleId, candle)
	m.meta.UpdateTime = time.Now().UnixMilli()
	return nil
}

func (m *Chart) GetUpdateTime() int64 {
	return m.meta.UpdateTime
}

func (m *Chart) indexing(interval string) int {
	return int(hash(interval)) % len(m.intervals)
}
