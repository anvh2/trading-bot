package server

import (
	"context"
	"runtime"
	"runtime/debug"

	"github.com/adshao/go-binance/v2/futures"
	"go.uber.org/zap"
)

func (s *Server) consume(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("[Consume] failed", zap.Any("error", r), zap.Any("debug", debug.Stack()))
		}
	}()

	listenKey, err := s.binance.GetListenKey(ctx)
	if err != nil {
		s.logger.Error("[Consume] failed to get listen key", zap.Error(err))
		return err
	}

	done, _, err := futures.WsUserDataServe(listenKey, func(event *futures.WsUserDataEvent) {
		switch event.Event {
		case futures.UserDataEventTypeOrderTradeUpdate:
			s.processOrder(ctx, event)

		case futures.UserDataEventTypeListenKeyExpired:
			s.logger.Info("[Consume] reconsume data", zap.Int("goroutines", runtime.NumGoroutine()))
			s.consume(ctx)
		}

	}, func(err error) {
		s.logger.Error("[Consume] failed to consume user data", zap.Error(err))
	})

	if err != nil {
		s.logger.Error("[Consume] failed to new user data stream", zap.Error(err))
		return err
	}

	<-done
	return nil
}
