package notify

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

type Polling func(idx int32)
type Process func(ctx context.Context, message *models.Chart) error

type NotifyWorker struct {
	logger  *logger.Logger
	process Process
	polling Polling
	message chan *models.Chart
	quit    chan struct{}
	wait    *sync.WaitGroup
	size    int32
}

func New(logger *logger.Logger, size int32) *NotifyWorker {
	return &NotifyWorker{
		logger:  logger,
		message: make(chan *models.Chart, size/4),
		quit:    make(chan struct{}),
		wait:    &sync.WaitGroup{},
		size:    size,
	}
}

func (w *NotifyWorker) WithPolling(polling Polling) {
	w.polling = polling
}

func (w *NotifyWorker) WithProcess(process Process) {
	w.process = process
}

func (w *NotifyWorker) Start() error {
	// start worker
	go func() {
		for i := int32(0); i < w.size; i++ {
			w.wait.Add(1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[NotifyWorker] process message failed", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				for {
					select {
					case msg := <-w.message:
						w.process(context.Background(), msg)

					case <-w.quit:
						return
					}
				}
			}()
		}
	}()

	// start poller
	go func() {
		size := w.size / 4
		if size == 0 {
			size = 1
		}

		for i := int32(0); i < size; i++ {
			w.wait.Add(1)

			ticker := time.NewTicker(time.Second)

			go func(idx int32) {
				defer func() {
					if r := recover(); r != nil {
						w.logger.Error("[NotifyWorker] failed to process", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
					}
				}()

				defer w.wait.Done()

				for {
					select {
					case <-ticker.C:
						w.polling(idx)

					case <-w.quit:
						return
					}
				}
			}(i)
		}
	}()

	return nil
}

func (w *NotifyWorker) Stop() {
	close(w.quit)
	w.wait.Wait()
}

func (w *NotifyWorker) SendJob(ctx context.Context, message *models.Chart) {
	w.message <- message
}
