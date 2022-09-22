package futures

import (
	"net/http"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/client"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/spf13/viper"
)

const (
	baseUrl    = "https://fapi.binance.com"
	testnetUrl = "https://testnet.binancefuture.com"
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
	baseUrl string
}

func New(logger *logger.Logger, binance *futures.Client, config *Config) *Futures {
	url := testnetUrl
	if viper.GetBool("binance.config.production") {
		url = baseUrl
	}

	return &Futures{
		config:  config,
		binance: binance,
		logger:  logger,
		client:  client.New(),
		baseUrl: url,
	}
}
