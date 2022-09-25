package trader

import (
	"context"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/cache/exchange"
	cachemock "github.com/anvh2/trading-bot/internal/cache/mocks"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/internal/service/binance"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProcessTrading(t *testing.T) {
	logger := logger.NewDev()
	viper.SetConfigFile("../../../config.dev.toml")
	viper.ReadInConfig()

	cases := []*struct {
		desc    string
		server  *Server
		message interface{}
	}{
		{
			desc: "happy case",
			server: &Server{
				logger:  logger,
				binance: binance.New(logger),
				supbot: func() *bot.TelegramBot {
					bot, err := bot.NewTelegramBot(logger, viper.GetString("telegram.trading_bot_token"))
					if err != nil {
						logger.Fatal("failed to new bot", zap.Error(err))
					}
					return bot
				}(),
				exchange: &cachemock.ExchangeMock{
					GetFunc: func(symbol string) (*exchange.Symbol, error) {
						return &exchange.Symbol{
							Filters: &exchange.Filters{
								{
									FilterType: futures.SymbolFilterTypePrice,
									TickSize:   "0.1",
								},
								{
									FilterType: futures.SymbolFilterTypeLotSize,
									StepSize:   "0.001",
								},
							},
						}, nil
					},
				},
			},
			message: &models.Oscillator{
				Symbol: "BTCUSDT",
				Stoch: map[string]*models.Stoch{
					"1h": {
						RSI: 80,
						K:   90,
						D:   90,
					},
				},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			err := test.server.ProcessTrading(context.Background(), test.message)
			assert.Nil(t, err)
		})
	}
}

func TestCalculateQuantity(t *testing.T) {
	quan := calculateQuantity(0.02216, 10)
	fmt.Println(quan)
}
