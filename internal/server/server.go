package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	supbot "github.com/anvh2/trading-bot/internal/bot/support-bot"
	"github.com/anvh2/trading-bot/internal/cache"
	cf "github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/crawler"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/worker"
	"github.com/go-redis/redis/v8"
)

type Server struct {
	logger  *logger.Logger
	config  *models.ExchangeConfig
	crawler *crawler.Crawler
	supbot  *supbot.TelegramBot
	cache   *cache.Cache
	storage *storage.Storage
	worker  *worker.Worker

	quitPolling chan struct{}
}

func NewServer(config *models.ExchangeConfig) *Server {
	logger, err := logger.New("./tmp/log.log")
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	redisCli := redis.NewClient(&redis.Options{
		Addr:       ":6379",
		DB:         1,
		MaxRetries: 5,
	})

	if err := redisCli.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to redis", err)
	}

	storage := storage.New(logger, redisCli)
	cache := cache.NewCache(cf.CandleLimit)

	supbot, err := supbot.NewTelegramBot(logger, "5629721774:AAH0Uq1xuqw7oKPSVQrNIDjeT8EgZgMuMZg")
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	crawler := crawler.New(logger, config, cache)

	server := &Server{
		logger:      logger,
		config:      config,
		supbot:      supbot,
		cache:       cache,
		storage:     storage,
		crawler:     crawler,
		quitPolling: make(chan struct{}),
	}

	server.supbot.Handle("/info", server.handleCommand)
	server.worker = worker.New(logger, 256, server.ProcessData)
	return server
}

func (s *Server) Start() error {
	s.crawler.Start()
	s.worker.Start()

	go s.polling()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		s.crawler.Stop()
		s.worker.Stop()
		close(s.quitPolling)
		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
