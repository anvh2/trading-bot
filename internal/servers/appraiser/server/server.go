package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anvh2/trading-bot/internal/logger"
	rpc_client "github.com/anvh2/trading-bot/internal/rpc/client"
	"github.com/anvh2/trading-bot/internal/services/binance"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/storage/order"
	"github.com/anvh2/trading-bot/internal/worker"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Server struct {
	logger   *logger.Logger
	order    storage.Order
	worker   *worker.Worker
	binance  *binance.Binance
	notifier notifier.NotifierServiceClient
}

func New() *Server {
	logger, err := logger.New(viper.GetString("appraiser.log_path"))
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	redisCli := redis.NewClient(&redis.Options{
		Addr:       viper.GetString("redis.addr"),
		DB:         1,
		MaxRetries: 5,
	})

	if err := redisCli.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to redis", zap.Error(err))
	}

	order := order.New(logger, redisCli)

	binance := binance.New(logger)

	worker, err := worker.New(logger, &worker.PoolConfig{NumPolling: 1})
	if err != nil {
		log.Fatal("failed to new worker", zap.Error(err))
	}

	conn, err := rpc_client.NewClient(viper.GetString("notifier.addr"), rpc_client.WithInsecure(), rpc_client.WithBlock())
	if err != nil {
		log.Fatal("failed to init notifier client conn", zap.Error(err))
	}

	notifier := notifier.NewNotifierServiceClient(conn)

	return &Server{
		order:    order,
		worker:   worker,
		logger:   logger,
		binance:  binance,
		notifier: notifier,
	}
}

func (s *Server) Start() error {
	go s.consume(context.Background())

	s.worker.WithPolling(s.appraise)
	s.worker.Start()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
