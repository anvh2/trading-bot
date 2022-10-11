package server

import (
	"context"
	"log"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/pubsub"
	rpc_client "github.com/anvh2/trading-bot/internal/rpc/client"
	rpc_server "github.com/anvh2/trading-bot/internal/rpc/server"
	"github.com/anvh2/trading-bot/internal/services/binance"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/worker"
	"github.com/anvh2/trading-bot/pkg/api/v1/analyzer"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	logger     *logger.Logger
	binance    *binance.Binance
	notifyDb   *storage.Notify
	subscriber pubsub.Subscriber
	publisher  pubsub.Publisher
	worker     *worker.Worker
	notifier   notifier.NotifierServiceClient
}

func New() *Server {
	logger, err := logger.New(viper.GetString("analyzer.log_path"))
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

	publisher := pubsub.New(logger, redisCli)
	subscriber := pubsub.New(logger, redisCli)
	notifyDb := storage.NewNotify(logger, redisCli)

	worker, err := worker.New(logger, &worker.PoolConfig{NumProcess: 64})
	if err != nil {
		log.Fatal("failed to new workder", zap.Error(err))
	}

	conn, err := rpc_client.NewClient(viper.GetString("notifier.addr"), rpc_client.WithInsecure(), rpc_client.WithBlock())
	if err != nil {
		log.Fatal("failed to init notifier client conn", zap.Error(err))
	}

	notifier := notifier.NewNotifierServiceClient(conn)
	binance := binance.New(logger)

	return &Server{
		logger:     logger,
		binance:    binance,
		notifyDb:   notifyDb,
		subscriber: subscriber,
		publisher:  publisher,
		worker:     worker,
		notifier:   notifier,
	}
}

func (s *Server) Start() error {
	s.worker.WithProcess(s.Process)

	s.subscriber.Subscribe(
		context.Background(),
		"trading.channel.analyze",
		func(ctx context.Context, message interface{}) error {
			s.worker.SendJob(ctx, message)
			return nil
		},
	)

	s.worker.Start()

	server := rpc_server.NewServer(
		viper.GetString("notifier.host"),
		viper.GetInt("notifier.port"),
		rpc_server.RegisterGRPCHandlerFunc(func(server *grpc.Server) {
			analyzer.RegisterAnalyzerServiceServer(server, s)
		}),
		rpc_server.WithShutdownHook(func() {
			s.notifyDb.Close()
			s.publisher.Close()
			s.subscriber.Close()
			s.worker.Stop()
		}),
	)

	return server.Start()
}
