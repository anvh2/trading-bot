package analyze

import (
	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/notify"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/trader"
)

type Server struct {
	supbot   *bot.TelegramBot
	market   *cache.Market
	database *storage.Notify
	analyze  *notify.NotifyWorker
	trader   *trader.Trader
}

func New(
	supbot *bot.TelegramBot,
	market *cache.Market,
	database *storage.Notify,
	analyze *notify.NotifyWorker,
	trader *trader.Trader,
) *Server {
	return &Server{
		supbot:   supbot,
		market:   market,
		database: database,
		analyze:  analyze,
		trader:   trader,
	}
}
