package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/cinar/indicator"
)

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

func TestCalculateStoch(t *testing.T) {
	// client := binance.NewClient("", "")
	// resp, err := client.NewKlinesService().Symbol("BTCUSDT").Interval("1d").Limit(int(1000)).Do(context.Background())
	// if err != nil {
	// 	return
	// }

	// low := make([]float64, len(resp))
	// high := make([]float64, len(resp))
	// close := make([]float64, len(resp))

	// for i := 0; i < len(resp); i++ {
	// 	l, _ := strconv.ParseFloat(resp[i].Low, 64)
	// 	low[i] = l

	// 	h, _ := strconv.ParseFloat(resp[i].High, 64)
	// 	high[i] = h

	// 	c, _ := strconv.ParseFloat(resp[i].Close, 64)
	// 	close[i] = c

	// }

	cli := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, "https://www.binance.com/fapi/v1/continuousKlines?limit=1000&pair=BTCUSDT&contractType=PERPETUAL&interval=1d", nil)
	if err != nil {
		fmt.Println("New Request Error", err)
		return
	}

	res, err := cli.Do(req)
	if err != nil {
		fmt.Println("Do Error", err)
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error read body", err)
		return
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			fmt.Println("Error close", err)
		}
	}()

	json, err := simplejson.NewJson(data)
	if err != nil {
		fmt.Println("error parse json", err)
		return
	}

	low := make([]float64, len(json.MustArray()))
	high := make([]float64, len(json.MustArray()))
	close := make([]float64, len(json.MustArray()))

	for i := 0; i < len(json.MustArray()); i++ {
		item := json.GetIndex(i)

		l, _ := strconv.ParseFloat(item.GetIndex(3).MustString(), 64)
		low[i] = l

		h, _ := strconv.ParseFloat(item.GetIndex(2).MustString(), 64)
		high[i] = h

		c, _ := strconv.ParseFloat(item.GetIndex(4).MustString(), 64)
		close[i] = c

	}

	k, d, j := indicator.Kdj(9, 3, 3, high, low, close)
	fmt.Print(k[len(k)-1], d[len(d)-1], j[len(j)-1], " ")
	_, rsi := indicator.RsiPeriod(14, close)
	fmt.Println(rsi[len(rsi)-1])
}
