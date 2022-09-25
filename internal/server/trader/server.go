package trader

import (
	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/service/binance"
)

type Server struct {
	logger   *logger.Logger
	supbot   *bot.TelegramBot
	binance  *binance.Binance
	exchange cache.Exchange
}

func New(
	logger *logger.Logger,
	exchange cache.Exchange,
	supbot *bot.TelegramBot,
	binance *binance.Binance,
) *Server {
	return &Server{
		logger:   logger,
		supbot:   supbot,
		binance:  binance,
		exchange: exchange,
	}
}
