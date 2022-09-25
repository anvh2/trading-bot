package server

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
	"github.com/anvh2/trading-bot/internal/cache/market"
	"github.com/anvh2/trading-bot/internal/crawler"
	"github.com/anvh2/trading-bot/internal/indicator"
	"github.com/anvh2/trading-bot/internal/logger"
	analyzeHandler "github.com/anvh2/trading-bot/internal/server/analyze"
	notifyHandler "github.com/anvh2/trading-bot/internal/server/notify"
	traderHandler "github.com/anvh2/trading-bot/internal/server/trader"
	"github.com/anvh2/trading-bot/internal/service/binance"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/internal/worker"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Server struct {
	logger   *logger.Logger
	crawler  *crawler.Crawler
	supbot   *bot.TelegramBot
	market   cache.Market
	exchange cache.Exchange
	notifyDb *storage.Notify
	analyze  *worker.Worker
	trader   *worker.Worker
	tradeDb  *storage.Position
	binance  *binance.Binance
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

	market := market.NewMarket(viper.GetInt32("chart.candles.limit"))
	exchange := exchange.New(logger)
	notifyDb := storage.NewNotify(logger, redisCli)
	tradeDb := storage.NewPosition(logger, redisCli)

	supbot, err := bot.NewTelegramBot(logger, viper.GetString("telegram.trading_bot_token"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	binance := binance.New(logger)
	crawler := crawler.New(logger, market, exchange, binance)

	analyze, err := worker.New(logger, &worker.PoolSize{Process: viper.GetInt32("notify.worker.size"), Polling: viper.GetInt32("notify.worker.size")})
	if err != nil {
		log.Fatal("failed to new notify worker", zap.Error(err))
	}

	trader, err := worker.New(logger, &worker.PoolSize{Process: viper.GetInt32("trader.size"), Polling: viper.GetInt32("trader.size")})
	if err != nil {
		log.Fatal("failed to new trading worker", zap.Error(err))
	}

	return &Server{
		logger:   logger,
		supbot:   supbot,
		market:   market,
		exchange: exchange,
		notifyDb: notifyDb,
		analyze:  analyze,
		tradeDb:  tradeDb,
		crawler:  crawler,
		binance:  binance,
		trader:   trader,
	}
}

func (s *Server) Setup() error {
	indicator.SetUp()

	notifySrv := notifyHandler.New(s.market)
	s.supbot.Handle("/info", notifySrv.ProcessAnalyzeCommand)

	analyzeHandler := analyzeHandler.New(
		s.supbot,
		s.market,
		s.exchange,
		s.notifyDb,
		s.analyze,
		s.trader,
	)

	s.analyze.WithPolling(analyzeHandler.NotifyPolling)
	s.analyze.WithProcess(analyzeHandler.ProcessNotify)

	traderHandler := traderHandler.New(
		s.logger,
		s.exchange,
		s.supbot,
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
		s.analyze.Start()
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
		s.analyze.Stop()
		s.trader.Stop()
		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
