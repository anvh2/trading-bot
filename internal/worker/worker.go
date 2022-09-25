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

type PoolSize struct {
	Process int32
	Polling int32
}

type Worker struct {
	logger  *logger.Logger
	process Process
	polling Polling
	message chan interface{}
	quit    chan struct{}
	wait    *sync.WaitGroup
	size    *PoolSize
}

func New(logger *logger.Logger, poolSize *PoolSize) (*Worker, error) {
	if poolSize == nil {
		return nil, errors.New("worker: poolsize invalid")
	}

	buffer := poolSize.Process / 2

	return &Worker{
		logger:  logger,
		message: make(chan interface{}, buffer),
		quit:    make(chan struct{}),
		wait:    &sync.WaitGroup{},
		size:    poolSize,
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
		for i := int32(0); i < w.size.Process; i++ {
			w.wait.Add(1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[Worker] process message failed", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				for {
					select {
					case msg := <-w.message:
						ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
						defer cancel()

						w.process(ctx, msg)

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
		for i := int32(0); i < w.size.Polling; i++ {
			w.wait.Add(1)

			go func(idx int32) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[Worker] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				ticker := time.NewTicker(time.Second)

				for {
					select {
					case <-ticker.C:
						ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
						defer cancel()

						w.polling(ctx, idx)

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
