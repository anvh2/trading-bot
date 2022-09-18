package futures

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-bot/internal/client"
	"github.com/bitly/go-simplejson"
)

func GetCandlesticks(symbol, interval string, limit int) ([]*binance.Kline, error) {
	url := fmt.Sprintf("https://www.binance.com/fapi/v1/continuousKlines?limit=%d&pair=%s&contractType=PERPETUAL&interval=%s", limit, symbol, interval)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	cli := client.New()

	res, err := cli.Do(req)
	if err != nil {
		return []*binance.Kline{}, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []*binance.Kline{}, err
	}
	defer func() {
		res.Body.Close()
	}()

	json, err := simplejson.NewJson(data)
	if err != nil {
		return []*binance.Kline{}, err
	}

	num := len(json.MustArray())
	resp := make([]*binance.Kline, num)
	for i := 0; i < num; i++ {
		item := json.GetIndex(i)
		if len(item.MustArray()) < 11 {
			return []*binance.Kline{}, fmt.Errorf("invalid kline response")
		}
		resp[i] = &binance.Kline{
			OpenTime:                 item.GetIndex(0).MustInt64(),
			Open:                     item.GetIndex(1).MustString(),
			High:                     item.GetIndex(2).MustString(),
			Low:                      item.GetIndex(3).MustString(),
			Close:                    item.GetIndex(4).MustString(),
			Volume:                   item.GetIndex(5).MustString(),
			CloseTime:                item.GetIndex(6).MustInt64(),
			QuoteAssetVolume:         item.GetIndex(7).MustString(),
			TradeNum:                 item.GetIndex(8).MustInt64(),
			TakerBuyBaseAssetVolume:  item.GetIndex(9).MustString(),
			TakerBuyQuoteAssetVolume: item.GetIndex(10).MustString(),
		}
	}

	return resp, nil
}
