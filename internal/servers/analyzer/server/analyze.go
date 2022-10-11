package server

import (
	"context"

	"github.com/anvh2/trading-bot/pkg/api/v1/analyzer"
)

func (s *Server) Analyze(ctx context.Context, req *analyzer.AnalyzeRequest) (*analyzer.AnalyzeResponse, error) {
	if err := validateAnalyzeRequest(req); req != nil {
		return nil, err
	}

	// for _, interval := range viper.GetStringSlice("market.intervals") {
	// 	resp, err := s.binance.ListCandlesticks(ctx, symbol, interval, viper.GetInt("chart.candles.limit"))
	// 	if err != nil {
	// 	}
	// }

	return nil, nil
}

func validateAnalyzeRequest(req *analyzer.AnalyzeRequest) error {
	return nil
}
