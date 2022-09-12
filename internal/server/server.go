package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anvh2/trading-bot/internal/crawler"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/anvh2/trading-bot/internal/service/notify"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/go-redis/redis/v8"
)

type Server struct {
	logger  *logger.Logger
	config  *models.BotConfig
	crawler *crawler.Crawler
	notify  *notify.TelegramBot
	storage *storage.Storage
}

func NewServer(config *models.BotConfig) *Server {
	logger, err := logger.New("./tmp/log.log")
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	redisCli := redis.NewClient(&redis.Options{
		Addr:       "0.0.0.0:6379",
		DB:         1,
		MaxRetries: 5,
	})

	if err := redisCli.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to redis", err)
	}

	storage := storage.New(logger, redisCli)

	notifyBot, err := notify.NewTelegramBot(logger, "5629721774:AAH0Uq1xuqw7oKPSVQrNIDjeT8EgZgMuMZg")
	if err != nil {
		log.Fatal("failed to new notify bot", err)
	}

	server := &Server{
		logger:  logger,
		config:  config,
		notify:  notifyBot,
		storage: storage,
	}

	server.crawler = crawler.New(logger, config.ExchangeConfig, server.ProcessCrawlerMessage)

	return server
}

func (s *Server) Start() error {
	s.crawler.Start()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		s.crawler.Stop()
		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
