package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/anvh2/trading-bot/internal/helpers"
	"github.com/anvh2/trading-bot/internal/indicator"
	logdev "github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	client "github.com/anvh2/trading-bot/internal/rpc/client"
	rpc "github.com/anvh2/trading-bot/internal/rpc/server"
	"github.com/anvh2/trading-bot/internal/storage"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	log       *logdev.Logger
	notifyDb  *storage.Notify
	notifyCli notifier.NotifierServiceClient
)

func TestWorker(t *testing.T) {
	log = logdev.NewDev()

	test_SetupServer()

	redisCli := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})

	if err := redisCli.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to new redis cli", zap.Error(err))
	}

	notifyDb = storage.NewNotify(log, redisCli)

	conn, err := client.NewClient(":8080", client.WithInsecure())
	if err != nil {
		log.Fatal("failed to new notify cli", zap.Error(err))
	}

	notifyCli = notifier.NewNotifierServiceClient(conn)

	w, _ := New(log, &PoolConfig{NumProcess: 64})
	w.WithProcess(test_Process)
	w.Start()

	go func() {
		ticker := time.NewTicker(time.Second)

		for range ticker.C {
			for i := 0; i < 1000; i++ {
				w.SendJob(context.Background(), func() string {
					chart := &models.Chart{
						Symbol: "BTCUSDT",
						// Candles: make(map[string][]*models.Candlestick),
					}
					return chart.String()
				}())
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)

		for range ticker.C {
			fmt.Println(runtime.NumGoroutine())
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		w.Stop()

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")
}

func test_Process(ctx context.Context, data interface{}) error {
	message := &models.Chart{}

	if err := json.Unmarshal([]byte(fmt.Sprint(data)), message); err != nil {
		log.Error("[Process] failed to unmarshal message", zap.Error(err))
		return err
	}

	oscillator := &models.Oscillator{
		Symbol: message.Symbol,
		Stoch:  make(map[string]*models.Stoch),
	}

	defer func() {
		message = nil
		oscillator.Stoch = nil
		runtime.GC()
	}()

	for interval, candles := range message.Candles {
		low := make([]float64, len(candles))
		high := make([]float64, len(candles))
		close := make([]float64, len(candles))

		for idx, candle := range candles {
			l, _ := strconv.ParseFloat(candle.Low, 64)
			low[idx] = l

			h, _ := strconv.ParseFloat(candle.High, 64)
			high[idx] = h

			c, _ := strconv.ParseFloat(candle.Close, 64)
			close[idx] = c
		}

		_, rsi := indicator.RSIPeriod(14, close)
		k, d, _ := indicator.KDJ(9, 3, 3, high, low, close)

		stoch := &models.Stoch{
			RSI: rsi[len(rsi)-1],
			K:   k[len(k)-1],
			D:   d[len(d)-1],
		}

		oscillator.Stoch[interval] = stoch
	}

	if !indicator.WithinRangeBound(oscillator.Stoch["1h"], indicator.RangeBoundRecommend) {
		return errors.New("analyze: not ready to trade")
	}

	msg := fmt.Sprintf("%s\t\t\t latest: -%0.4f(s)\n\t%s\n", message.Symbol, float64((time.Now().UnixMilli()-message.UpdateTime))/1000.0, helpers.ResolvePositionSide(oscillator.GetRSI()))

	for _, interval := range viper.GetStringSlice("market.intervals") {
		stoch, ok := oscillator.Stoch[interval]
		if !ok {
			log.Error("[Process] stoch in interval invalid", zap.Any("stoch", stoch))
			return errors.New("analyze: stoch in interval invalid")
		}

		msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.K, stoch.D)
	}

	if err := notifyDb.Create(ctx, message.Symbol); err != nil {
		return err
	}

	channel := cast.ToString(viper.GetInt64("notify.channels.futures_recommendation"))
	_, err := notifyCli.Push(ctx, &notifier.PushRequest{Channel: channel, Message: msg})
	if err != nil {
		log.Error("[Process] failed to push notification", zap.String("channel", channel), zap.Error(err))
	}

	return nil
}

type test_Handler struct{}

func (h *test_Handler) Push(ctx context.Context, req *notifier.PushRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func test_SetupServer() {
	h := &test_Handler{}

	server := rpc.NewServer(
		"localhost",
		8080,
		rpc.RegisterGRPCHandlerFunc(func(server *grpc.Server) {
			notifier.RegisterNotifierServiceServer(server, h)
		}),
	)

	go server.Start()
}
