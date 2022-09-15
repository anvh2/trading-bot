package cache

import (
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/cache/circular"
	"github.com/anvh2/trading-bot/internal/models"
)

type Market struct {
	mutex   *sync.RWMutex
	symbol  string
	candles map[string]*circular.Cache // map[interval]candles
	meta    *models.MarketMetadata
	limit   int32 // limit of candles's length
}

func (m *Market) Init(symbol string, limit int32) *Market {
	return &Market{
		mutex:   &sync.RWMutex{},
		symbol:  symbol,
		candles: make(map[string]*circular.Cache),
		meta: &models.MarketMetadata{
			UpdateTime: time.Hour.Milliseconds(),
		},
		limit: limit,
	}
}

func (m *Market) Candles(interval string) *circular.Cache {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.candles[interval]
}

func (m *Market) CreateCandle(interval string, candle *models.Candlestick) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.candles[interval] == nil {
		m.candles[interval] = circular.New(m.limit)
	}

	m.candles[interval].Create(candle)
	m.meta.UpdateTime = time.Now().UnixMilli()
}

func (m *Market) Metadata() *models.MarketMetadata {
	return m.meta
}

func (m *Market) UpdateMeta() {
	m.meta.UpdateTime = time.Hour.Milliseconds()
}
