package futures

import (
	"net/http"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/client"
	"github.com/anvh2/trading-bot/internal/logger"
)

type Config struct {
	ApiKey    string
	SecretKey string
}

type Futures struct {
	config  *Config
	binance *futures.Client
	logger  *logger.Logger
	client  *http.Client
}

func New(logger *logger.Logger, binance *futures.Client, config *Config) *Futures {
	return &Futures{
		config:  config,
		binance: binance,
		logger:  logger,
		client:  client.New(),
	}
}
