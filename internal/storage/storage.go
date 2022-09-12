package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anvh2/trading-bot/internal/logger"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	expirationOscillator time.Duration = 10 * time.Minute
)

var (
	ErrOscillatorIsAlreadyExist error = errors.New("oscillator is already exist")
)

var (
	keyOscillatorData string = "trading_boy.oscillator_data.%s"
)

type Storage struct {
	logger *logger.Logger
	db     *redis.Client
}

func New(logger *logger.Logger, db *redis.Client) *Storage {
	return &Storage{
		db:     db,
		logger: logger,
	}
}

func (s *Storage) SetNXOscillator(ctx context.Context, data *models.Oscillator) error {
	key := fmt.Sprintf(keyOscillatorData, data.Symbol)
	effected, err := s.db.SetNX(ctx, key, data.String(), expirationOscillator).Result()
	if err != nil {
		s.logger.Error("[Storage][SetNXOscillator] failed to set oscillator to redis", zap.Any("data", data), zap.Error(err))
		return err
	}

	if !effected {
		s.logger.Error("[Storage][SetNXOscillator] oscillator is alreay exist to redis", zap.Any("data", data))
		return ErrOscillatorIsAlreadyExist
	}

	s.logger.Info("[Storage][SetNXOscillator] set oscillator success", zap.Any("data", data))
	return nil
}
