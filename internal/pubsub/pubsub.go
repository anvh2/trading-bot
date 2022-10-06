package pubsub

import (
	"context"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type PubSub struct {
	logger *logger.Logger
	db     *redis.Client
	quit   chan struct{}
}

func New(logger *logger.Logger, db *redis.Client) *PubSub {
	return &PubSub{
		logger: logger,
		db:     db,
		quit:   make(chan struct{}),
	}
}

func (pb *PubSub) Publish(ctx context.Context, channel string, data interface{}) error {
	_, err := pb.db.Publish(ctx, channel, data).Result()
	if err != nil {
		pb.logger.Error("[Publish] failed", zap.Error(err))
		return err
	}

	return nil
}

func (pb *PubSub) Subscribe(ctx context.Context, channel string, process Process) {
	subscriber := pb.db.Subscribe(ctx, channel)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pb.logger.Error("[Subcribe] failed to process", zap.Any("error", r))
			}
		}()

		for {
			select {
			case <-pb.quit:
				return

			default:
				msg, err := subscriber.ReceiveMessage(ctx)
				if err != nil {
					pb.logger.Error("[Subscribe] failed to receive message", zap.Error(err))
					continue
				}

				process(ctx, msg.Payload)
			}
		}
	}()
}

func (pb *PubSub) Close() {
	close(pb.quit)
	pb.db.Close()
}
