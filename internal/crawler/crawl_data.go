package crawler

import (
	"runtime/debug"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-bot/internal/cache"
	"github.com/anvh2/trading-bot/internal/config"
	"github.com/anvh2/trading-bot/internal/models"
	"go.uber.org/zap"
)

func (c *Crawler) CrawlData() error {
	for _, interval := range config.Intervals {
		pair := make(map[string]string, len(c.market.Symbols()))
		for _, symbol := range c.market.Symbols() {
			pair[symbol] = interval
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error("[Crawler][CrawlData] failed to start, recovered", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			done, _, err := binance.WsCombinedKlineServe(pair, c.handleKlinesStreamData, c.handleKlinesStreamError)
			if err != nil {
				c.logger.Error("[Crawler][CrawlData] failed to connect to klines stream data", zap.Error(err))
				return
			}

			<-done
		}()

		time.Sleep(2 * time.Second)
	}

	return nil
}

func (c *Crawler) handleKlinesStreamData(event *binance.WsKlineEvent) {
	chart, err := c.market.Chart(event.Symbol)
	if err == cache.ErrorChartNotFound {
		chart = c.market.CreateChart(event.Symbol)
	}

	candles, err := chart.Candles(event.Kline.Interval)
	if err == cache.ErrorCandlesNotFound {
		return
	}

	last, idx := candles.Last()
	if idx < 0 {
		return
	}

	lastCandle, ok := last.(*models.Candlestick)
	if !ok {
		return
	}

	// update the last candle
	if lastCandle.OpenTime == event.Kline.StartTime &&
		lastCandle.CloseTime == event.Kline.EndTime {

		lastCandle.Close = event.Kline.Close
		lastCandle.High = event.Kline.High
		lastCandle.Low = event.Kline.Low

		chart.UpdateCandle(event.Kline.Interval, idx, lastCandle)
		return
	}

	// create new candle
	candle := &models.Candlestick{
		OpenTime:  event.Kline.StartTime,
		CloseTime: event.Kline.EndTime,
		Low:       event.Kline.Low,
		High:      event.Kline.High,
		Close:     event.Kline.Close,
	}

	chart.CreateCandle(event.Kline.Interval, candle)
}

func (c *Crawler) handleKlinesStreamError(err error) {
	c.logger.Error("[Crawler][CrawlData] failed to recieve stream data", zap.Error(err))
}
