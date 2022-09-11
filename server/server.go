package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anvh2/trading-boy/binance"
	"github.com/anvh2/trading-boy/logger"
	"github.com/anvh2/trading-boy/models"
)

type Server struct {
	config   *models.BotConfig
	exchange binance.Exchange
}

func NewServer(config *models.BotConfig) *Server {
	logger, err := logger.New("./tmp/log.log")
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	exchange := binance.NewBinanceWrapper(
		logger,
		config.ExchangeConfig.PublicKey,
		config.ExchangeConfig.SecretKey,
		config.ExchangeConfig.DepositAddresses,
	)

	return &Server{
		config:   config,
		exchange: exchange,
	}
}

func (s *Server) Start() error {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		// run hooks here
		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
