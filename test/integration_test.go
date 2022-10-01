package test

import (
	"context"
	"fmt"
	"log"
	"testing"

	rpc "github.com/anvh2/trading-bot/internal/rpc/client"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	"go.uber.org/zap"
)

func TestPushNotification(t *testing.T) {
	conn, err := rpc.NewClient(":5500", rpc.WithInsecure())
	if err != nil {
		log.Fatal("failed to new connection", zap.Error(err))
	}

	client := notifier.NewNotifierServiceClient(conn)

	resp, err := client.Push(context.Background(), &notifier.PushRequest{Channel: "-1", Message: "hello world"})
	if err != nil {
		log.Fatal("failed to call api", zap.Error(err))
	}

	fmt.Println("Success", resp)
}
