// package main

// import (
// 	bot "github.com/anvh2/trading-boy/cmd"
// )

// const (
// 	version = "0.0.1-pre-alpha"
// )

// func main() {
// 	bot.SetVersion(version)
// 	bot.Execute()
// }

package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/markcheno/go-talib"
)

var (

	//Binance
	binanceAPIKey    = "<your binance api key>"
	binanceSecretKey = "<your binance secret key>"
)

func humanTimeToUnixNanoTime(input time.Time) int64 {
	return int64(time.Nanosecond) * input.UnixNano() / int64(time.Millisecond)
}

func main() {

	client := binance.NewClient(binanceAPIKey, binanceSecretKey)

	targetSymbol := "BTCUSDT"
	targetInterval := "30m"

	data, err := client.NewKlinesService().Symbol(targetSymbol).Interval(targetInterval).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(len(data))

	inputs := []float64{}
	for _, e := range data {
		input, _ := strconv.ParseFloat(e.Close, 64)
		inputs = append(inputs, input)
	}

	rsi := talib.Rsi(inputs, 14)
	log.Println("RSI : ", rsi[len(rsi)-1])
}
