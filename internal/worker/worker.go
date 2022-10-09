package worker

import (
	"context"
	"errors"
	"runtime/debug"
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/logger"
	"go.uber.org/zap"
)

type Polling func(ctx context.Context, idx int32)
type Process func(ctx context.Context, message interface{}) error

type PoolConfig struct {
	NumProcess int32
	NumPolling int32
}

type Worker struct {
	logger  *logger.Logger
	process Process
	polling Polling
	message chan interface{}
	quit    chan struct{}
	wait    *sync.WaitGroup
	config  *PoolConfig
}

func New(logger *logger.Logger, config *PoolConfig) (*Worker, error) {
	if config == nil {
		return nil, errors.New("worker: config invalid")
	}

	buffer := config.NumProcess / 2

	return &Worker{
		logger:  logger,
		message: make(chan interface{}, buffer),
		quit:    make(chan struct{}),
		wait:    &sync.WaitGroup{},
		config:  config,
	}, nil
}

func (w *Worker) WithPolling(polling Polling) {
	w.polling = polling
}

func (w *Worker) WithProcess(process Process) {
	w.process = process
}

func (w *Worker) Start() error {
	// start worker
	go func() {
		for i := int32(0); i < w.config.NumProcess; i++ {
			w.wait.Add(1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[Worker] process message failed", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				for {
					select {
					case msg, ok := <-w.message:
						if !ok {
							continue
						}

						if w.process != nil {
							w.process(ctx, msg)
						}

					case <-w.quit:
						if len(w.message) == 0 {
							return
						}
					}
				}
			}()
		}
	}()

	// start poller
	go func() {
		for i := int32(0); i < w.config.NumPolling; i++ {
			w.wait.Add(1)

			go func(idx int32) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[Worker] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				ticker := time.NewTicker(time.Second)

				for {
					select {
					case _, ok := <-ticker.C:
						if !ok {
							continue
						}

						if w.polling != nil {
							w.polling(ctx, idx)
						}

					case <-w.quit:
						return
					}
				}
			}(i)
		}
	}()

	return nil
}

func (w *Worker) Stop() {
	close(w.quit)
	w.wait.Wait()
	close(w.message)
}

func (w *Worker) SendJob(ctx context.Context, message interface{}) {
	w.message <- message
}
