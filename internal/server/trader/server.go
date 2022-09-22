package trader

import (
	binance "github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/service/futures"
)

type Server struct {
	logger  *logger.Logger
	supbot  *bot.TelegramBot
	futures *futures.Futures
	binance *binance.Client
}

func New(
	logger *logger.Logger,
	supbot *bot.TelegramBot,
	futures *futures.Futures,
	binance *binance.Client,
) *Server {
	return &Server{
		logger:  logger,
		supbot:  supbot,
		futures: futures,
		binance: binance,
	}
}
