package server

import (
	"context"

	"go.uber.org/zap"
)

const (
	candlesLookback = 5 // 5 hours
)

func (s *Server) appraise(ctx context.Context, idx int32) error {
	symbol, err := s.order.PopQueue(ctx)
	if err != nil {
		return err
	}

	candles, err := s.binance.ListCandlesticks(ctx, symbol, "1h", candlesLookback)
	if err != nil {
		s.logger.Error("[Appraise] failed to get candles", zap.String("symbol", symbol), zap.Error(err))
		return err
	}

	// appraize candles in sideway
	for _, candle := range candles {
		s.logger.Info("[Appraise] candle", zap.Any("candle", candle))
	}

	return nil
}
