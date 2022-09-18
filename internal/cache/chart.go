package cache

import (
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/cache/circular"
	"github.com/anvh2/trading-bot/internal/models"
)

type ChartMeta struct {
	UpdateTime int64
}

type Chart struct {
	mutex  *sync.RWMutex
	symbol string                     // key of chart
	cache  map[string]*circular.Cache // map[interval]candles
	meta   *ChartMeta
	limit  int32 // limit of candles's length
}

func (m *Chart) Init(symbol string, limit int32) *Chart {
	return &Chart{
		mutex:  &sync.RWMutex{},
		symbol: symbol,
		cache:  make(map[string]*circular.Cache),
		meta:   &ChartMeta{},
		limit:  limit,
	}
}

func (m *Chart) Candles(interval string) (*circular.Cache, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.cache[interval] == nil {
		return nil, ErrorCandlesNotFound
	}

	return m.cache[interval], nil
}

func (m *Chart) CreateCandle(interval string, candle *models.Candlestick) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.cache[interval] == nil {
		m.cache[interval] = circular.New(m.limit)
	}

	m.cache[interval].Create(candle)
	m.meta.UpdateTime = time.Now().UnixMilli()

	return nil
}

func (m *Chart) UpdateCandle(interval string, candleId int32, candle *models.Candlestick) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.cache[interval] == nil {
		return ErrorCandlesNotFound
	}

	m.cache[interval].Update(candleId, candle)
	m.meta.UpdateTime = time.Now().UnixMilli()
	return nil
}

func (m *Chart) GetUpdateTime() int64 {
	return m.meta.UpdateTime
}
