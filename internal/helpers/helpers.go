package helpers

import "github.com/anvh2/trading-bot/internal/models"

func CheckIndicatorIsReadyToTrade(oscillator *models.Oscillator) bool {
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
