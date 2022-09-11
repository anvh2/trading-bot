package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/markcheno/go-talib"
)

func (s *Server) ProcessCrawlerMessage(ctx context.Context, message []interface{}) error {
	inputs := []float64{}
	for _, e := range message {
		input, _ := strconv.ParseFloat(fmt.Sprint(e), 64)
		inputs = append(inputs, input)
	}

	rsi := talib.Rsi(inputs, 14)

	msg := fmt.Sprintf("RSI BTCUSDT-1m: %v\n ", rsi[len(rsi)-1])
	s.notify.Push(ctx, 1630847448, msg)

	return nil
}
