package notify

import (
	"github.com/anvh2/trading-bot/internal/cache"
)

type Server struct {
	market cache.Market
}

func New(market cache.Market) *Server {
	return &Server{
		market: market,
	}
}
