package storage

import (
	"context"
	"fmt"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	keyOrders   = "trading_bot.orders.%d"
	keyPosition = "trading_bot.position.%s"
)

type Position struct {
	logger *logger.Logger
	db     *redis.Client
}

func NewPosition(logger *logger.Logger, db *redis.Client) *Position {
	return &Position{
		logger: logger,
		db:     db,
	}
}

func (db *Position) CreatePosition(ctx context.Context, position *models.Position) error {
	key := fmt.Sprintf(keyPosition, position.Symbol)
	res, err := db.db.HSet(ctx, key, position.PositionSide, position.String()).Result()
	if err != nil {
		db.logger.Error("[Storage][CreatePosition] failed to create", zap.Any("position", position), zap.Error(err))
		return err
	}

	db.logger.Info("[Storage][CreatePosition] success", zap.Any("position", position), zap.Int64("res", res))
	return nil
}

func (db *Position) ReadPosition(ctx context.Context, symbol string, side models.PositionSide) (*models.Position, error) {
	position := &models.Position{}

	key := fmt.Sprintf(keyPosition, symbol)
	res, err := db.db.HGet(ctx, key, string(side)).Result()
	if err != nil {
		db.logger.Error("[Storage][ReadPosition] failed to read", zap.String("symbol", symbol), zap.Any("side", side), zap.Error(err))
		return position, err
	}

	if err := position.Parse(res); err != nil {
		db.logger.Error("[Storage][ReadPosition] failed to parse", zap.String("symbol", symbol), zap.Any("side", side), zap.String("res", res), zap.Error(err))
		return position, err
	}

	db.logger.Info("[Storage][ReadPosition] success", zap.String("symbol", symbol), zap.Any("side", side), zap.Any("position", position))
	return position, nil
}

func (db *Position) ReadAllPositions(ctx context.Context, symbol string) ([]*models.Position, error) {
	res, err := db.db.HGetAll(ctx, fmt.Sprintf(keyPosition, symbol)).Result()
	if err != nil {
		db.logger.Error("[Storage][ReadAllPositions] failed to read", zap.String("symbol", symbol), zap.Error(err))
		return []*models.Position{}, err
	}

	idx := 0
	postions := make([]*models.Position, len(res))

	for side, data := range res {
		if err := postions[idx].Parse(data); err != nil {
			db.logger.Error("[Storage][ReadAllPositions] failed to parse", zap.String("symbol", symbol), zap.Any("res", res), zap.String("side", side), zap.Error(err))
			return postions, err
		}
	}

	db.logger.Info("[Storage][ReadAllPositions] success", zap.String("symbol", symbol), zap.Any("positions", postions))
	return postions, nil
}

func (db *Position) CreateOrder(ctx context.Context, order *models.Order) error {
	key := fmt.Sprintf(keyOrders, order.Symbol)

	res, err := db.db.HSet(ctx, key, "order.OrderId", order.String()).Result()
	if err != nil {
		db.logger.Error("[Storage][CreateOrder] failed to create", zap.Any("order", order), zap.Error(err))
		return err
	}

	db.logger.Info("[Storage][CreateOrder] success", zap.Any("order", order), zap.Int64("res", res))
	return nil
}

func (db *Position) ReadOrder(ctx context.Context, symbol string, orderId int64) (*models.Order, error) {
	order := &models.Order{}

	key := fmt.Sprintf(keyOrders, symbol)
	res, err := db.db.HGet(ctx, key, fmt.Sprint(orderId)).Result()
	if err != nil {
		db.logger.Error("[Storage][ReadOrder] failed to read", zap.String("symbol", symbol), zap.Int64("orderId", orderId), zap.Error(err))
		return order, err
	}

	if err := order.Parse(res); err != nil {
		db.logger.Error("[Storage][ReadOrder] failed to parse", zap.String("symbol", symbol), zap.Int64("orderId", orderId), zap.String("res", res), zap.Error(err))
		return order, err
	}

	db.logger.Info("[Storage][ReadOrder] success", zap.String("symbol", symbol), zap.Int64("orderId", orderId), zap.Any("order", order))
	return order, nil
}

func (db *Position) ReadAllOrders(ctx context.Context, symbol string) ([]*models.Order, error) {
	key := fmt.Sprintf(keyOrders, symbol)

	res, err := db.db.HGetAll(ctx, key).Result()
	if err != nil {
		db.logger.Error("[Storage][ReadAllOrders] failed to read", zap.String("symbol", symbol), zap.Error(err))
		return []*models.Order{}, err
	}

	idx := 0
	orders := make([]*models.Order, len(res))

	for orderId, orderData := range res {
		if err := orders[idx].Parse(orderData); err != nil {
			db.logger.Error("[Storage][ReadAllOrders] failed to parse", zap.String("symbol", symbol), zap.String("orderId", orderId), zap.Any("res", res), zap.Error(err))
			return orders, err
		}
		idx++
	}

	db.logger.Info("[Storage][ReadAllOrders] success", zap.String("symbol", symbol), zap.Any("orders", orders))
	return orders, nil
}
