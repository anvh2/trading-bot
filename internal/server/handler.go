package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/anvh2/trading-boy/internal/crawler"
	"github.com/markcheno/go-talib"
	"go.uber.org/zap"
)

var (
	focusedInterval = map[string]bool{"5m": true, "15m": true, "30m": true, "1h": true, "4h": true, "1d": true}
)

type ReviewData struct {
	RSI float64
}

func (s *Server) ProcessCrawlerMessage(ctx context.Context, message *crawler.Message) error {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("[ProcessCrawlerMessage] process message failed", zap.Any("error", r))
		}
	}()

	rsis := map[string]float64{}

	for interval, prices := range message.Prices {
		if !focusedInterval[interval] {
			continue
		}

		inputs := []float64{}
		for _, price := range prices {
			input, _ := strconv.ParseFloat(price, 64)
			inputs = append(inputs, input)
		}

		result := talib.Rsi(inputs, 14)
		rsi := result[len(result)-1]

		rsis[interval] = rsi
	}

	if !reviewOK(rsis) {
		return nil
	}

	msg := fmt.Sprintf("%s\n", message.Symbol)

	for interval, rsi := range rsis {
		msg += fmt.Sprintf("%s: %v\n", interval, rsi)

	}

	s.notify.Push(ctx, 1630847448, msg)

	return nil
}

func reviewOK(rsis map[string]float64) bool {
	counter := 0
	for _, rsi := range rsis {
		if rsi == 0 {
			counter++
			continue
		}

		if rsi < 70 && rsi > 30 {
			counter++
			continue
		}
	}

	return counter <= 1
}
