package server

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/indicator"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

func (s *Server) ProcessData(ctx context.Context, message *models.Chart) error {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("[ProcessData] process message failed", zap.String("symbol", message.Symbol), zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
		}
	}()

	if message == nil ||
		message.Candles == nil ||
		len(message.Candles) == 0 {
		return errors.New("message invalid")
	}

	oscillator := &models.Oscillator{
		Symbol: message.Symbol,
		Stoch:  make(map[string]*models.Stoch),
	}

	for interval, candles := range message.Candles {
		low := make([]float64, len(candles))
		high := make([]float64, len(candles))
		close := make([]float64, len(candles))

		for idx, candle := range candles {
			l, _ := strconv.ParseFloat(candle.Low, 64)
			low[idx] = l

			h, _ := strconv.ParseFloat(candle.High, 64)
			high[idx] = h

			c, _ := strconv.ParseFloat(candle.Close, 64)
			close[idx] = c
		}

		_, rsi := indicator.RSIPeriod(14, close)
		k, d, _ := indicator.KDJ(9, 3, 3, high, low, close)

		stoch := &models.Stoch{
			RSI: rsi[len(rsi)-1],
			K:   k[len(k)-1],
			D:   d[len(d)-1],
		}

		oscillator.Stoch[interval] = stoch
	}

	if !isReadyToTrade(oscillator) {
		return errors.New("not ready to trade")
	}

	msg := fmt.Sprintf("%s\t\t\t latency: +%0.4f(s)\n", message.Symbol, float64((time.Now().UnixMilli()-message.UpdateTime)/1000))

	for _, interval := range config.Intervals {
		stoch, ok := oscillator.Stoch[interval]
		if !ok {
			return errors.New("stoch in interval invalid")
		}

		msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.K, stoch.D)
	}

	if err := s.notify.Create(ctx, message.Symbol); err != nil {
		return err
	}

	return s.supbot.PushNotify(ctx, config.TelegramChatId, msg)
}

func isReadyToTrade(oscillator *models.Oscillator) bool {
	stoch := oscillator.Stoch["1h"]
	if stoch == nil {
		return false
	}

	if stoch.RSI >= 70 || stoch.RSI <= 30 {
		if (stoch.K >= 80 || stoch.K <= 20) &&
			(stoch.D >= 80 || stoch.D <= 20) {
			return true
		}
	}
	return false
}
