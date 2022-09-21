package trader

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

type Polling func(ctx context.Context)
type Process func(ctx context.Context, message *models.Oscillator) error

type Trader struct {
	logger   *logger.Logger
	notify   chan *models.Oscillator
	binance  *futures.Client
	poolSize int32
	wait     *sync.WaitGroup
	quit     chan struct{}
	process  Process
	polling  Polling
}

func New(logger *logger.Logger, poolSize int32, binance *futures.Client, process Process, polling Polling) *Trader {
	return &Trader{
		logger:   logger,
		notify:   make(chan *models.Oscillator),
		binance:  binance,
		poolSize: poolSize,
		wait:     &sync.WaitGroup{},
		quit:     make(chan struct{}),
		process:  process,
		polling:  polling,
	}
}

func (t *Trader) Start() {
	// start trader
	go func() {
		for i := int32(0); i < t.poolSize; i++ {
			t.wait.Add(1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.logger.Error("[Trader] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer t.wait.Done()

				for {
					select {
					case message := <-t.notify:
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						t.process(ctx, message)

					case <-t.quit:
						return
					}
				}
			}()
		}
	}()

	// start poller
	go func() {
		for i := int32(0); i < t.poolSize; i++ {
			t.wait.Add(1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.logger.Error("[Trader] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer t.wait.Done()

				ticker := time.NewTicker(time.Second)

				for {
					select {
					case <-ticker.C:
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						t.polling(ctx)

					case <-t.quit:
						return
					}
				}
			}()
		}
	}()
}

func (t *Trader) Stop() {
	close(t.quit)
	t.wait.Wait()
}

func (t *Trader) SendNotify(ctx context.Context, notify *models.Oscillator) {
	t.notify <- notify
}
