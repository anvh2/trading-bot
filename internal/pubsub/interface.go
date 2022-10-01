package pubsub

import "context"

type Process func(ctx context.Context, message interface{}) error

type Publisher interface {
	Publish(ctx context.Context, channel string, data interface{}) error
	Close()
}

type Subscriber interface {
	Subscribe(ctx context.Context, channel string, process Process)
	Close()
}
