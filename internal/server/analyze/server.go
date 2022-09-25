package analyze

import (
	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/worker"
)

type Server struct {
	supbot   *bot.TelegramBot
	market   cache.Market
	exchange cache.Exchange
	database *storage.Notify
	analyze  *worker.Worker
	trader   *worker.Worker
}

func New(
	supbot *bot.TelegramBot,
	market cache.Market,
	exchange cache.Exchange,
	database *storage.Notify,
	analyze *worker.Worker,
	trader *worker.Worker,
) *Server {
	return &Server{
		supbot:   supbot,
		market:   market,
		exchange: exchange,
		database: database,
		analyze:  analyze,
		trader:   trader,
	}
}
