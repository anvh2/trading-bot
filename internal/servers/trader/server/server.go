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
	"github.com/anvh2/trading-bot/internal/services/binance"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

type Server struct {
	logger    *logger.Logger
	binance   *binance.Binance
	exchange  cache.Exchange
	subcriber pubsub.Subscriber
}

func New() *Server {
	logger, err := logger.New(viper.GetString("trader.log_path"))
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

	exchange := exchange.New(logger)

	binance := binance.New(logger)

	subciber := pubsub.New(logger, redisCli)

	return &Server{
		logger:    logger,
		binance:   binance,
		exchange:  exchange,
		subcriber: subciber,
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
