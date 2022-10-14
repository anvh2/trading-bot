package order

import (
	"context"
	"fmt"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	keyOrders = "trading_bot.orders.%d"
)

type Order struct {
	logger *logger.Logger
	db     *redis.Client
}

func New(logger *logger.Logger, db *redis.Client) *Order {
	return &Order{
		db:     db,
		logger: logger,
	}
}

func (db *Order) Set(ctx context.Context, order *models.Order) error {
	key := fmt.Sprintf(keyOrders, order.Symbol)

	res, err := db.db.HSet(ctx, key, "order.OrderId", order.String()).Result()
	if err != nil {
		db.logger.Error("[Order][Set] failed", zap.Any("order", order), zap.Error(err))
		return err
	}

	db.logger.Info("[Order][Set] success", zap.Any("order", order), zap.Int64("res", res))
	return nil
}

func (db *Order) MSet(ctx context.Context, symbol string, orders ...*models.Order) error {
	key := fmt.Sprintf(keyOrders, symbol)

	data := make(map[string]interface{})

	for _, order := range orders {
		data[order.OrderId] = order.String()
	}

	effected, err := db.db.HMSet(ctx, key, data).Result()
	if err != nil {
		db.logger.Error("[Order][MSet] failed", zap.Any("orders", orders), zap.Error(err))
		return err
	}

	db.logger.Info("[Order][MSet] success", zap.Any("orders", orders), zap.Bool("effected", effected))
	return nil
}

func (db *Order) Get(ctx context.Context, symbol string, orderId int64) (*models.Order, error) {
	order := &models.Order{}

	key := fmt.Sprintf(keyOrders, symbol)
	res, err := db.db.HGet(ctx, key, fmt.Sprint(orderId)).Result()
	if err != nil {
		db.logger.Error("[Order][Get] failed to get", zap.String("symbol", symbol), zap.Int64("orderId", orderId), zap.Error(err))
		return order, err
	}

	if err := order.Parse(res); err != nil {
		db.logger.Error("[Order][Get] failed to parse", zap.String("symbol", symbol), zap.Int64("orderId", orderId), zap.String("res", res), zap.Error(err))
		return order, err
	}

	db.logger.Info("[Order][Get] success", zap.String("symbol", symbol), zap.Int64("orderId", orderId), zap.Any("order", order))
	return order, nil
}

func (db *Order) GetAll(ctx context.Context, symbol string) ([]*models.Order, error) {
	key := fmt.Sprintf(keyOrders, symbol)

	res, err := db.db.HGetAll(ctx, key).Result()
	if err != nil {
		db.logger.Error("[Order][GetAll] failed to get", zap.String("symbol", symbol), zap.Error(err))
		return []*models.Order{}, err
	}

	idx := 0
	orders := make([]*models.Order, len(res))

	for orderId, orderData := range res {
		if err := orders[idx].Parse(orderData); err != nil {
			db.logger.Error("[Order][GetAll] failed to parse", zap.String("symbol", symbol), zap.String("orderId", orderId), zap.Any("res", res), zap.Error(err))
			return orders, err
		}
		idx++
	}

	db.logger.Info("[Order][GetAll] success", zap.String("symbol", symbol), zap.Any("orders", orders))
	return orders, nil
}

func (db *Order) Exists(ctx context.Context, symbol string) bool {
	key := fmt.Sprintf(keyOrders, symbol)

	exists, err := db.db.Exists(ctx, key).Result()
	if err != nil {
		db.logger.Error("[Order][Exists] failed to check", zap.String("symbol", symbol), zap.Error(err))
		return false
	}

	db.logger.Info("[Order][Exists] success to check", zap.String("symbol", symbol), zap.Int64("exists", exists))
	return exists > 0
}

func (db *Order) Close() {
	db.db.Close()
}
