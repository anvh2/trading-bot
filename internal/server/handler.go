package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/anvh2/trading-bot/internal/crawler"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/markcheno/go-talib"
	"go.uber.org/zap"
)

const (
	inTimePeriod int = 14
)

const (
	groupChatId int64 = -653827904
)

var (
	sortedInterval  = []string{"5m", "15m", "30m", "1h", "4h", "1d"}
	focusedInterval = map[string]bool{"5m": true, "15m": true, "30m": true, "1h": true, "4h": true, "1d": true}
)

func (s *Server) ProcessCrawlerMessage(ctx context.Context, message *crawler.Message) error {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("[ProcessCrawlerMessage] process message failed", zap.Any("error", r))
		}
	}()

	if len(message.CandleSticks) == 0 {
		return nil
	}

	oscillator := &models.Oscillator{
		Symbol: message.Symbol,
		Stoch:  make(map[string]*models.Stoch),
	}

	for interval, candles := range message.CandleSticks {
		if !focusedInterval[interval] {
			continue
		}

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

		slowK, slowD := talib.Stoch(inHight, inLow, inClose, 12, 3, talib.SMA, 3, talib.SMA)
		result := talib.Rsi(inClose, inTimePeriod)
		rsi := result[len(result)-1]

		stoch := &models.Stoch{
			RSI:   rsi,
			SlowK: slowK[len(slowK)-1],
			SlowD: slowD[len(slowD)-1],
		}

		oscillator.Stoch[interval] = stoch
	}

	if !isReadyToTrade(oscillator) {
		return nil
	}

	if err := s.storage.SetNXOscillator(ctx, oscillator); err != nil {
		s.logger.Info("[ProcessCrawlerMessage] already send notification", zap.Any("oscillator", oscillator), zap.Error(err))
		return nil
	}

	msg := fmt.Sprintf("%s\n", message.Symbol)

	for _, interval := range sortedInterval {
		stoch := oscillator.Stoch[interval]
		msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.SlowK, stoch.SlowD)
	}

	return s.notify.Push(ctx, groupChatId, msg)
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
