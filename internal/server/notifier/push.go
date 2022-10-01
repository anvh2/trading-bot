package notifier

import (
	"context"

	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) Push(ctx context.Context, req *notifier.PushRequest) (*emptypb.Empty, error) {
	err := s.notify.PushNotify(ctx, cast.ToInt64(req.Channel), req.Message)
	if err != nil {
		s.logger.Error("[Push] failed", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	s.logger.Info("[Push] success", zap.Any("req", req))
	return &emptypb.Empty{}, nil
}
