package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	binance "github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/bot"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/crawler"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/notify"
	analyzeHandler "github.com/anvh2/trading-bot/internal/server/analyze"
	notifyHandler "github.com/anvh2/trading-bot/internal/server/notify"
	traderHandler "github.com/anvh2/trading-bot/internal/server/trader"
	"github.com/anvh2/trading-bot/internal/service/futures"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/trader"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

type Server struct {
	logger   *logger.Logger
	crawler  *crawler.Crawler
	supbot   *bot.TelegramBot
	market   *cache.Market
	notifyDb *storage.Notify
	notify   *notify.NotifyWorker
	trader   *trader.Trader
	tradeDb  *storage.Position
	futures  *futures.Futures
	binance  *binance.Client
}

func NewServer() *Server {
	logger, err := logger.New(viper.GetString("server.log_path"))
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

	market := cache.NewMarket(viper.GetInt32("chart.candles.limit"))
	notifyDb := storage.NewNotify(logger, redisCli)
	tradeDb := storage.NewPosition(logger, redisCli)

	supbot, err := bot.NewTelegramBot(logger, viper.GetString("telegram.trading_bot_token"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	binance := binance.NewClient(viper.GetString("binance.config.api_key"), viper.GetString("binance.config.secret_key"))
	futures := futures.New(logger, binance, &futures.Config{ApiKey: viper.GetString("binance.config.api_key"), SecretKey: viper.GetString("binance.config.secret_key")})
	crawler := crawler.New(logger, market, binance, futures)
	notify := notify.New(logger, viper.GetInt32("notify_worker.size"))
	trader := trader.New(logger, viper.GetInt32("trader.size"), binance)

	return &Server{
		logger:   logger,
		supbot:   supbot,
		market:   market,
		notifyDb: notifyDb,
		notify:   notify,
		tradeDb:  tradeDb,
		crawler:  crawler,
		futures:  futures,
		binance:  binance,
		trader:   trader,
	}
}

func (s *Server) Setup() error {
	notifySrv := notifyHandler.New(s.market)
	s.supbot.Handle("/info", notifySrv.ProcessAnalyzeCommand)

	analyzeHandler := analyzeHandler.New(
		s.supbot,
		s.market,
		s.notifyDb,
		s.notify,
		s.trader,
	)

	s.notify.WithPolling(analyzeHandler.NotifyPolling)
	s.notify.WithProcess(analyzeHandler.ProcessNotify)

	traderHandler := traderHandler.New(
		s.logger,
		s.supbot,
		s.futures,
		s.binance,
	)

	s.trader.WithPolling(traderHandler.TraderPolling)
	s.trader.WithProcess(traderHandler.ProcessTrading)

	return nil
}

func (s *Server) Start() error {
	ready := s.crawler.Start()

	go func() {
		<-ready
		s.notify.Start()
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
		s.notify.Stop()
		s.trader.Stop()
		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
