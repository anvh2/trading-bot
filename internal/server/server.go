package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anvh2/trading-boy/internal/crawler"
	"github.com/anvh2/trading-boy/internal/logger"
	"github.com/anvh2/trading-boy/internal/models"
	"github.com/anvh2/trading-boy/internal/service/notify"
	"go.uber.org/zap"
)

type Server struct {
	config  *models.BotConfig
	crawler *crawler.Crawler
	notify  *notify.TelegramBot
}

func NewServer(config *models.BotConfig) *Server {
	logger, err := logger.New("./tmp/log.log")
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	notifyBot, err := notify.NewTelegramBot(logger, "5629721774:AAH0Uq1xuqw7oKPSVQrNIDjeT8EgZgMuMZg")
	if err != nil {
		logger.Fatal("failed to new notify bot", zap.Error(err))
	}

	server := &Server{
		config: config,
		notify: notifyBot,
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
