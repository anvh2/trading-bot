package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	ftbinc "github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/cache"
	cf "github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/crawler"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/notify"
	"github.com/anvh2/trading-bot/internal/service/futures"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/trader"
	"github.com/go-redis/redis/v8"
)

type Server struct {
	logger   *logger.Logger
	crawler  *crawler.Crawler
	supbot   *bot.TelegramBot
	market   *cache.Market
	notifyDb *storage.Notify
	notifyWr *notify.NotifyWorker
	trader   *trader.Trader
	tradeDb  *storage.Position
	futures  *futures.Futures
	binance  *ftbinc.Client
}

func NewServer() *Server {
	logger, err := logger.New("./logs/server.log")
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

	market := cache.NewMarket(cf.Intervals, cf.CandleLimit)
	notifyDb := storage.NewNotify(logger, redisCli)
	tradeDb := storage.NewPosition(logger, redisCli)

	supbot, err := bot.NewTelegramBot(logger, os.Getenv("TELE_BOT_TOKEN"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	binance := ftbinc.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY"))
	futures := futures.New(logger, binance, &futures.Config{ApiKey: os.Getenv("BINANCE_API_KEY"), SecretKey: os.Getenv("BINANCE_SECRET_KEY")})
	crawler := crawler.New(logger, market, binance, futures)

	server := &Server{
		logger:   logger,
		supbot:   supbot,
		market:   market,
		notifyDb: notifyDb,
		tradeDb:  tradeDb,
		crawler:  crawler,
		futures:  futures,
		binance:  binance,
	}

	server.supbot.Handle("/info", server.handleCommand)
	server.notifyWr = notify.New(logger, 64, server.ProcessNotify, server.NotifyPolling)
	server.trader = trader.New(logger, 2, binance, server.ProcessTrading, server.TraderPolling)

	return server
}

func (s *Server) Start() error {
	ready := s.crawler.Start()

	go func() {
		<-ready
		s.notifyWr.Start()
		s.trader.Start()
		s.logger.Info("[Start] start success and ready to trade")
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		s.crawler.Stop()
		s.notifyWr.Stop()
		s.trader.Stop()
		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
