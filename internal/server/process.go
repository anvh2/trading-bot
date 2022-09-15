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
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/markcheno/go-talib"
	"go.uber.org/zap"
)

func (s *Server) ProcessData(ctx context.Context, message *models.CandlesMarket) error {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("[ProcessData] process message failed", zap.String("symbol", message.Symbol), zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
		}
	}()

	if message == nil ||
		message.Candlesticks == nil ||
		len(message.Candlesticks) == 0 {
		return errors.New("message invalid")
	}

	oscillator := &models.Oscillator{
		Symbol: message.Symbol,
		Stoch:  make(map[string]*models.Stoch),
	}

	for interval, candles := range message.Candlesticks {
		inLow := make([]float64, len(candles))
		inHight := make([]float64, len(candles))
		inClose := make([]float64, len(candles))

		for idx, candle := range candles {
			low, _ := strconv.ParseFloat(candle.Low, 64)
			inLow[idx] = low

			hight, _ := strconv.ParseFloat(candle.High, 64)
			inHight[idx] = hight

			close, _ := strconv.ParseFloat(candle.Close, 64)
			inClose[idx] = close
		}

		stochSettings := config.StochSettings[config.StochShortTerm]

		slowK, slowD := talib.Stoch(inHight, inLow, inClose, stochSettings.FastKPeriod, stochSettings.SlowKPeriod, talib.SMA, stochSettings.SlowDPeriod, talib.SMA)
		result := talib.Rsi(inClose, config.RSIPeriod)
		rsi := result[len(result)-1]

		stoch := &models.Stoch{
			RSI:   rsi,
			SlowK: slowK[len(slowK)-1],
			SlowD: slowD[len(slowD)-1],
		}

		oscillator.Stoch[interval] = stoch
	}

	if !isReadyToTrade(oscillator) {
		return errors.New("not ready to trade")
	}

	msg := fmt.Sprintf("%s\t\t latency: +%d(ms)\n", message.Symbol, time.Now().UnixMilli()-message.UpdateTime)

	for _, interval := range config.Intervals {
		stoch, ok := oscillator.Stoch[interval]
		if !ok {
			return errors.New("stoch in interval invalid")
		}

		msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.SlowK, stoch.SlowD)
	}

	if err := s.storage.SetNXOscillator(ctx, oscillator); err != nil {
		return err
	}

	return s.notify.Push(ctx, config.TelegramChatId, msg)
}

func isReadyToTrade(oscillator *models.Oscillator) bool {
	counter := 0
	for _, stoch := range oscillator.Stoch {
		if stoch.RSI == 0 {
			counter++
			continue
		}

		if stoch.RSI < 70 && stoch.RSI > 30 {
			counter++
			continue
		}
	}

	return counter <= 1
}
