package trader

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/cache/exchange"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/pubsub"
	"github.com/anvh2/trading-bot/internal/service/binance"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

type Server struct {
	logger    *logger.Logger
	binance   *binance.Binance
	exchange  cache.Exchange
	notifyBot *bot.TelegramBot
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

	notifyBot, err := bot.NewTelegramBot(logger, viper.GetString("telegram.trading_bot_token"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	binance := binance.New(logger)

	subciber := pubsub.New(logger, redisCli)

	return &Server{
		logger:    logger,
		notifyBot: notifyBot,
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
		s.notifyBot.Stop()
		s.subcriber.Close()

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
