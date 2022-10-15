package market

import (
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/cache/circular"
	"github.com/anvh2/trading-bot/internal/cache/errors"
	"github.com/anvh2/trading-bot/internal/models"
)

type Chart struct {
	mutex  *sync.RWMutex
	symbol string                           // key of chart
	cache  map[string]*circular.Cache       // map[interval]candles
	meta   map[string]*models.ChartMetadata // map[internval]meta
	limit  int32                            // limit of candles's length
}

func (m *Chart) Init(symbol string, limit int32) *Chart {
	return &Chart{
		mutex:  &sync.RWMutex{},
		symbol: symbol,
		cache:  make(map[string]*circular.Cache),
		meta:   make(map[string]*models.ChartMetadata),
		limit:  limit,
	}
}

func (m *Chart) Candles(interval string) (*circular.Cache, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.cache[interval] == nil {
		return nil, errors.ErrorCandlesNotFound
	}

	return m.cache[interval], nil
}

func (m *Chart) CreateCandle(interval string, candle *models.Candlestick) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.cache[interval] == nil {
		m.cache[interval] = circular.New(m.limit)
		m.meta[interval] = &models.ChartMetadata{
			UpdateTime: time.Now().UnixMilli(),
		}
	}

	m.cache[interval].Create(candle)
	return nil
}

func (m *Chart) UpdateCandle(interval string, candleId int32, candle *models.Candlestick) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.cache[interval] == nil {
		m.cache[interval] = circular.New(m.limit)
		m.meta[interval] = &models.ChartMetadata{}
	}

	m.cache[interval].Update(candleId, candle)
	m.meta[interval].UpdateTime = time.Now().UnixMilli()
}

func (m *Chart) GetMetadata(interval string) *models.ChartMetadata {
	if m.meta != nil && m.meta[interval] != nil {
		return m.meta[interval]
	}
	return &models.ChartMetadata{}
}
