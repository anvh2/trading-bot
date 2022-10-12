package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/cache/exchange"
	cachemock "github.com/anvh2/trading-bot/internal/cache/mocks"
	"github.com/anvh2/trading-bot/internal/helpers"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/internal/services/binance"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	ntfmock "github.com/anvh2/trading-bot/pkg/api/v1/notifier/mock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestProcess(t *testing.T) {
	logger := logger.NewDev()
	viper.SetConfigFile("../config.toml")
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
				notifier: &ntfmock.NotifierServiceClientMock{
					PushFunc: func(ctx context.Context, in *notifier.PushRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
						return nil, nil
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
			err := test.server.Process(context.Background(), test.message)
			assert.Nil(t, err)
		})
	}
}

func TestCalculateQuantity(t *testing.T) {
	raw := calculateQuantity(19135.50, 45)
	fmt.Println(raw)

	quantity := helpers.AlignQuantity(raw, "0.001")
	fmt.Println(quantity)
}
