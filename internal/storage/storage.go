package storage

import (
	"context"

	"github.com/anvh2/trading-bot/internal/models"
)

//go:generate moq -pkg storagemock -out ./mocks/storage.mock.go . Notify
type Notify interface {
	Create(ctx context.Context, notifyId string) error
	Close()
}

//go:generate moq -pkg storagemock -out ./mocks/storage.mock.go . Order
type Order interface {
	Set(ctx context.Context, order *models.Order) error
	MSet(ctx context.Context, symbol string, orders ...*models.Order) error
	Get(ctx context.Context, symbol string, orderId int64) (*models.Order, error)
	GetAll(ctx context.Context, symbol string) ([]*models.Order, error)
	Exists(ctx context.Context, symbol string) bool
	AddQueue(ctx context.Context, symbol string) error
	RemoveQueue(ctx context.Context, symbol string) error
	PopQueue(ctx context.Context) (string, error)
	Close()
}
