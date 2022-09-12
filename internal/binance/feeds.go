package binance

import (
	"context"

	"github.com/anvh2/trading-bot/internal/models"
	"github.com/shopspring/decimal"
)

func (bw *BinanceWrapper) GetCandles(ctx context.Context, market *models.Market) ([]models.Candlestick, error) {
	binanceCandles, err := bw.api.NewKlinesService().Interval("30m").Symbol(market.MarketName).Do(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]models.Candlestick, len(binanceCandles))

	for i, binanceCandle := range binanceCandles {
		ret[i] = models.Candlestick{
			High:   binanceCandle.High,
			Open:   binanceCandle.Open,
			Close:  binanceCandle.Close,
			Low:    binanceCandle.Low,
			Volume: binanceCandle.Volume,
		}
	}

	return ret, nil
}

func (bw *BinanceWrapper) GetMarketSummary(ctx context.Context, market *models.Market) (*models.MarketSummary, error) {
	binanceSummary, err := bw.api.NewListPriceChangeStatsService().Symbol(market.MarketName).Do(ctx)
	if err != nil {
		return nil, err
	}

	ask, _ := decimal.NewFromString(binanceSummary[0].AskPrice)
	bid, _ := decimal.NewFromString(binanceSummary[0].BidPrice)
	high, _ := decimal.NewFromString(binanceSummary[0].HighPrice)
	low, _ := decimal.NewFromString(binanceSummary[0].LowPrice)
	volume, _ := decimal.NewFromString(binanceSummary[0].Volume)

	return &models.MarketSummary{
		Last:   ask,
		Ask:    ask,
		Bid:    bid,
		High:   high,
		Low:    low,
		Volume: volume,
	}, nil
}

func (bw *BinanceWrapper) GetOrderBook(ctx context.Context, market *models.Market) (*models.OrderBook, error) {
	orderbook, _, err := bw.orderbookFromREST(ctx, market)
	if err != nil {
		return nil, err
	}

	return orderbook, nil
}

func (wrapper *BinanceWrapper) orderbookFromREST(ctx context.Context, market *models.Market) (*models.OrderBook, int64, error) {
	binanceOrderBook, err := wrapper.api.NewDepthService().Symbol(market.MarketName).Do(ctx)
	if err != nil {
		return nil, -1, err
	}

	var orderBook models.OrderBook

	for _, ask := range binanceOrderBook.Asks {
		qty, err := decimal.NewFromString(ask.Quantity)
		if err != nil {
			return nil, -1, err
		}

		value, err := decimal.NewFromString(ask.Price)
		if err != nil {
			return nil, -1, err
		}

		orderBook.Asks = append(orderBook.Asks, models.Order{
			Quantity: qty,
			Value:    value,
		})
	}

	for _, bid := range binanceOrderBook.Bids {
		qty, err := decimal.NewFromString(bid.Quantity)
		if err != nil {
			return nil, -1, err
		}

		value, err := decimal.NewFromString(bid.Price)
		if err != nil {
			return nil, -1, err
		}

		orderBook.Bids = append(orderBook.Bids, models.Order{
			Quantity: qty,
			Value:    value,
		})
	}

	return &orderBook, binanceOrderBook.LastUpdateID, nil
}
