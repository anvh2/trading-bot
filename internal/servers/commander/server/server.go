package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/cache/exchange"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/pubsub"
	rpc_client "github.com/anvh2/trading-bot/internal/rpc/client"
	"github.com/anvh2/trading-bot/internal/services/binance"
	storage "github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/storage/order"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Server struct {
	logger    *logger.Logger
	order     storage.Order
	binance   *binance.Binance
	exchange  cache.Exchange
	subcriber pubsub.Subscriber
	notifier  notifier.NotifierServiceClient
}

func New() *Server {
	logger, err := logger.New(viper.GetString("commander.log_path"))
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	redisCli := redis.NewClient(&redis.Options{
		Addr:       viper.GetString("redis.addr"),
		DB:         1,
		MaxRetries: 5,
	})

	if err := redisCli.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to redis", err)
	}

	orderDb := order.New(logger, redisCli)

	exchange := exchange.New(logger)

	binance := binance.New(logger)

	subciber := pubsub.New(logger, redisCli)

	conn, err := rpc_client.NewClient(viper.GetString("notifier.addr"), rpc_client.WithInsecure(), rpc_client.WithBlock())
	if err != nil {
		log.Fatal("failed to init notifier client conn", zap.Error(err))
	}

	notifier := notifier.NewNotifierServiceClient(conn)

	return &Server{
		logger:    logger,
		order:     orderDb,
		binance:   binance,
		exchange:  exchange,
		subcriber: subciber,
		notifier:  notifier,
	}
}

func (s *Server) Start() error {
	s.subcriber.Subscribe(context.Background(), "trading.channel.trading", s.Process)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		s.subcriber.Close()

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
