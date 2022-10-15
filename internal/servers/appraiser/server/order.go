package server

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Server) processOrder(ctx context.Context, event *futures.WsUserDataEvent) {
	s.logger.Info("[ProcessOrder] event", zap.Any("event", event))

	// appraise for new position filled
	order := event.OrderTradeUpdate
	if order.Status == futures.OrderStatusTypeFilled || order.Status == futures.OrderStatusTypePartiallyFilled {
		if order.Type == futures.OrderTypeLimit || order.Type == futures.OrderTypeMarket {
			s.order.AddQueue(ctx, order.Symbol)
		}

		if order.Type == futures.OrderTypeTakeProfit || order.Type == futures.OrderTypeTakeProfitMarket {
			s.order.RemoveQueue(ctx, order.Symbol)
		}

		if order.Type == futures.OrderTypeStop || order.Type == futures.OrderTypeStopMarket {
			s.order.RemoveQueue(ctx, order.Symbol)
		}
	}

	// if order.Status == futures.OrderStatusTypePartiallyFilled {

	// }

	// if order.Status == futures.OrderStatusTypeCanceled {

	// }

	// notify
	channel := viper.GetInt64("notify.channels.futures_recommendation")
	msg := fmt.Sprintf("%s %s: %s | Price: %s | Quantity: %s | Status: %s", order.PositionSide, order.Symbol, order.Side, order.StopPrice, order.OriginalQty, order.Status)
	_, err := s.notifier.Push(ctx, &notifier.PushRequest{Channel: cast.ToString(channel), Message: msg})
	if err != nil {
		s.logger.Error("[ProcessOrder] failed to push notification", zap.Int64("channel", channel), zap.String("message", msg), zap.Error(err))
		return
	}

	s.logger.Info("[ProcessOrder] success", zap.String("msg", msg))
}
