package cache

import (
	"fmt"
	"testing"

	"github.com/anvh2/trading-bot/internal/config"
)

// var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func TestHash(t *testing.T) {
	// randStringRunes := func(n int) string {
	// 	b := make([]rune, n)
	// 	for i := range b {
	// 		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	// 	}
	// 	return string(b)
	// }

	data := map[int32]string{}

	for _, interval := range config.Intervals {
		idx := int32(hash(interval))
		if val, ok := data[idx]; ok {
			fmt.Println(val, interval)
			return
		}
		data[idx] = interval
	}

	fmt.Println(data)

	// mux := &sync.Mutex{}
	// wg := &sync.WaitGroup{}

	// for i := 0; i < 1; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		for k := 0; k < 100; k++ {
	// 			mux.Lock()
	// 			defer mux.Unlock()

	// 			str := randStringRunes(10)
	// 			idx := hash(str)
	// 			if val, ok := data[idx]; ok {
	// 				fmt.Println(val, str)
	// 				return
	// 			}
	// 			data[idx] = str
	// 		}
	// 	}()
	// }

	// wg.Wait()
}
