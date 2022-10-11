package server

import (
	"log"

	rpc "github.com/anvh2/trading-bot/internal/rpc/server"
	"github.com/anvh2/trading-bot/internal/services/telegram"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"google.golang.org/grpc"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/spf13/viper"
)

type Server struct {
	logger *logger.Logger
	notify *telegram.TelegramBot
}

func New() *Server {
	logger, err := logger.New(viper.GetString("notifier.log_path"))
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	notify, err := telegram.NewTelegramBot(logger, viper.GetString("telegram.trading_bot_token"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	return &Server{
		logger: logger,
		notify: notify,
	}
}

func (s *Server) Start() error {
	server := rpc.NewServer(
		viper.GetString("notifier.host"),
		viper.GetInt("notifier.port"),
		rpc.RegisterGRPCHandlerFunc(func(server *grpc.Server) {
			notifier.RegisterNotifierServiceServer(server, s)
		}),
		rpc.WithShutdownHook(func() {
			s.notify.Stop()
		}),
	)

	return server.Start()
}
