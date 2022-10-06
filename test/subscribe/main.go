package main

import (
	"context"
	"fmt"
	"time"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/pubsub"
	"github.com/go-redis/redis/v8"
)

func main() {
	redisCli := redis.NewClient(&redis.Options{
		Addr:       "159.223.67.54:6379",
		DB:         1,
		MaxRetries: 5,
	})

	subscriber := pubsub.New(logger.NewDev(), redisCli)

	subscriber.Subscribe(
		context.Background(),
		"trading.channel.analyze",
		func(ctx context.Context, message interface{}) error {
			fmt.Println(message)
			return nil
		},
	)

	time.Sleep(time.Minute)
}
