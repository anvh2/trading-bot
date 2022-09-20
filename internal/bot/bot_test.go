package bot

import (
	"context"
	"fmt"
	"testing"

	"github.com/anvh2/trading-bot/internal/logger"
)

func TestSend(t *testing.T) {
	bot, err := NewTelegramBot(logger.NewDev(), "5629721774:AAH0Uq1xuqw7oKPSVQrNIDjeT8EgZgMuMZg")
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.PushNotify(context.Background(), -653827904, "hello world")
}
