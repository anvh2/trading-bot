package server

import (
	"context"
	"testing"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/services/binance"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestProcessOrderStatus(t *testing.T) {
	logger := logger.NewDev()

	viper.SetConfigFile("../config.toml")
	viper.ReadInConfig()

	godotenv.Load("../../../../.env")

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
			},
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			err := test.server.consume(context.Background())
			assert.Nil(t, err)
		})
	}
}
