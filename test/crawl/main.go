package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	limiter := rate.NewLimiter(rate.Every(time.Minute), 1000)

	client := &http.Client{}
	wg := &sync.WaitGroup{}
	start := time.Now()

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 500; j++ {
				limiter.Wait(context.Background())

				url := "https://www.binance.com/fapi/v1/continuousKlines?limit=1000&pair=BTCUSDT&contractType=PERPETUAL&interval=1h"

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					return
				}

				res, err := client.Do(req)
				if err != nil {
					return
				}

				fmt.Println(res.Status)
			}
		}()
	}

	wg.Wait()
	fmt.Println(time.Since(start).Seconds())
}
