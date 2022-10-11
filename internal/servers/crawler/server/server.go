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
	"github.com/anvh2/trading-bot/internal/cache/market"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/internal/pubsub"
	"github.com/anvh2/trading-bot/internal/services/binance"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var (
	blacklist = map[string]bool{}
)

type Server struct {
	logger    *logger.Logger
	binance   *binance.Binance
	market    cache.Market
	exchange  cache.Exchange
	message   chan *models.Chart
	retryCh   chan *models.Symbol
	publisher pubsub.Publisher
	quit      chan struct{}
}

func New() *Server {
	logger, err := logger.New(viper.GetString("crawler.log_path"))
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

	market := market.NewMarket(viper.GetInt32("chart.candles.limit"))

	exchange := exchange.New(logger)
	binance := binance.New(logger)
	publisher := pubsub.New(logger, redisCli)

	return &Server{
		logger:    logger,
		binance:   binance,
		market:    market,
		exchange:  exchange,
		publisher: publisher,
		retryCh:   make(chan *models.Symbol, 100),
		quit:      make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ready := make(chan bool)

	go func() {
		s.retry()
		s.crawl()
		s.consume()
		ready <- true
	}()

	go func() {
		<-ready
		s.produce()
		s.refresh()
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		s.publisher.Close()

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
