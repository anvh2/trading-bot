package analyze

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anvh2/trading-bot/internal/helpers"
	"github.com/anvh2/trading-bot/internal/indicator"
	"github.com/anvh2/trading-bot/internal/models"
	"github.com/spf13/viper"
)

func (s *Server) ProcessNotify(ctx context.Context, message *models.Chart) error {
	if err := validateNotifyMessage(message); err != nil {
		return err
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

	if !helpers.CheckIndicatorIsReadyToTrade(oscillator) {
		return errors.New("not ready to trade")
	}

	msg := fmt.Sprintf("%s\t\t\t latency: +%0.4f(s)\n", message.Symbol, float64((time.Now().UnixMilli()-message.UpdateTime))/1000.0)

	for _, interval := range viper.GetStringSlice("market.intervals") {
		stoch, ok := oscillator.Stoch[interval]
		if !ok {
			return errors.New("stoch in interval invalid")
		}

		msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.K, stoch.D)
	}

	if err := s.database.Create(ctx, message.Symbol); err != nil {
		return err
	}

	s.trader.SendNotify(ctx, oscillator)
	s.supbot.PushNotify(ctx, viper.GetInt64("notify_channels.futures_recommendation"), msg)

	return nil
}

func validateNotifyMessage(message *models.Chart) error {
	if message == nil {
		return errors.New("notify: message invalid")
	}
	return nil
}
