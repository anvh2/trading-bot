package binance

import (
	"net/http"

	"github.com/anvh2/trading-bot/internal/client"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

type Binance struct {
	limiter *rate.Limiter
	logger  *logger.Logger
	client  *http.Client
}

func New(logger *logger.Logger) *Binance {
	limiter := rate.NewLimiter(
		rate.Every(viper.GetDuration("binance.rate_limit.duration")),
		viper.GetInt("binance.rate_limit.requests"),
	)
	return &Binance{
		limiter: limiter,
		logger:  logger,
		client:  client.New(),
	}
}
