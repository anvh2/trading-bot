package crawler

// import (
// 	"fmt"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"testing"

// 	"github.com/anvh2/trading-bot/internal/logger"
// 	"github.com/anvh2/trading-bot/internal/models"
// )

// func TestCrawl(t *testing.T) {
// 	logger, _ := logger.New("../../tmp/log.log")
// 	crawler := New(logger, &models.ExchangeConfig{}, nil, nil)
// 	crawler.symbols = []string{"BTCUSDT", "ETHUSDT"}
// 	crawler.Start()

// 	sigs := make(chan os.Signal, 1)
// 	done := make(chan bool)
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

// 	fmt.Println("Server now listening")

// 	go func() {
// 		<-sigs
// 		// run hooks here
// 		close(done)
// 	}()

// 	fmt.Println("Ctrl-C to interrupt...")
// 	<-done
// 	fmt.Println("Exiting...")
// }
