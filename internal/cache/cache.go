package cache

import (
	"github.com/anvh2/trading-bot/internal/cache/exchange"
	"github.com/anvh2/trading-bot/internal/cache/market"
)

//go:generate moq -pkg cachemock -out ./mocks/market_mock.go . Market
type Market interface {
	Chart(symbol string) (*market.Chart, error)
	CreateChart(symbol string) *market.Chart
	UpdateChart(symbol string) *market.Chart
}

//go:generate moq -pkg cachemock -out ./mocks/exchange_mock.go . Exchange
type Exchange interface {
	Set(symbols []*exchange.Symbol)
	Get(symbol string) (*exchange.Symbol, error)
	Symbols() []string
}
