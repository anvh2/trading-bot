package tradebot

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/logger"
)

// crawl position and cache
// crawl order and cache

// buy short long (4 cmd and backoff 3%)

// stop loss and take profit

//

type TradeBot struct {
	logger  *logger.Logger
	notify  chan interface{}
	binance *futures.Client
}

func New(logger *logger.Logger) *TradeBot {
	binance := futures.NewClient("tshhh50wl5HeGOkDXuA4soO81AWyX3AztDb9KoedzZuQ1CSpVidXllJAJzPhXGUB", "KGzctvmH5tsAm4GMTKxbVMwPFybnqIgWBH2rtVgalwyJpM1H2Qax7hyvnYH5i8hy")
	return &TradeBot{
		logger:  logger,
		binance: binance,
	}
}

func (t *TradeBot) Start() {

}

func (t *TradeBot) SendNotify(ctx context.Context, notify interface{}) {
	t.notify <- notify
}
