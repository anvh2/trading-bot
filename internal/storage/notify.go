package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	expirationNotify time.Duration = 10 * time.Minute
)

var (
	ErrNotifyIsAlreadyExist error = errors.New("notify: is already exist")
)

var (
	keyNotify string = "trading_bot.notify.%s"
)

type Notify struct {
	logger *logger.Logger
	db     *redis.Client
}

func NewNotify(logger *logger.Logger, db *redis.Client) *Notify {
	return &Notify{
		db:     db,
		logger: logger,
	}
}

func (s *Notify) Create(ctx context.Context, notifyId string) error {
	key := fmt.Sprintf(keyNotify, notifyId)

	effected, err := s.db.SetNX(ctx, key, "", expirationNotify).Result()
	if err != nil {
		s.logger.Error("[Notify][Cache] failed to set oscillator to redis", zap.Any("notifyId", notifyId), zap.Error(err))
		return err
	}

	if !effected {
		return ErrNotifyIsAlreadyExist
	}

	s.logger.Info("[Notify][Cache] set oscillator success", zap.Any("notifyId", notifyId))
	return nil
}
