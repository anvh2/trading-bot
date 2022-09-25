package binance

import (
	"net/http"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/client"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

type Binance struct {
	limiter *rate.Limiter
	futures *futures.Client
	logger  *logger.Logger
	client  *http.Client
}

func New(logger *logger.Logger) *Binance {
	futures := futures.NewClient(viper.GetString("binance.config.exchange.api_key"), viper.GetString("binance.config.exchange.secret_key"))

	limiter := rate.NewLimiter(
		rate.Every(viper.GetDuration("binance.rate_limit.duration")),
		viper.GetInt("binance.rate_limit.requests"),
	)
	return &Binance{
		limiter: limiter,
		futures: futures,
		logger:  logger,
		client:  client.New(),
	}
}
