package worker

import (
	"context"
	"sync"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
)

type Process func(ctx context.Context, message *models.CandlestickChart) error

type Worker struct {
	logger  *logger.Logger
	process Process
	message chan *models.CandlestickChart
	quit    chan struct{}
	wait    *sync.WaitGroup
	size    int32
}

func New(logger *logger.Logger, size int32, process Process) *Worker {
	return &Worker{
		logger:  logger,
		process: process,
		message: make(chan *models.CandlestickChart),
		quit:    make(chan struct{}),
		wait:    &sync.WaitGroup{},
		size:    size,
	}
}

func (w *Worker) Start() error {
	go func() {
		for i := int32(0); i < w.size; i++ {
			w.wait.Add(1)

			go func() {
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

	return nil
}

func (w *Worker) Stop() {
	close(w.quit)
	w.wait.Wait()
}

func (w *Worker) SendJob(ctx context.Context, message *models.CandlestickChart) {
	w.message <- message
}
