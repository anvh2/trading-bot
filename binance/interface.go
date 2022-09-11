package binance

import (
	"context"

	"github.com/anvh2/trading-boy/models"
	"github.com/shopspring/decimal"
)

// TradeType represents a type of order, from trading fees point of view.
type TradeType string

const (
	// TakerTrade represents the "buy" order type.
	TakerTrade = "taker"
	// MakerTrade represents the "sell" order type.
	MakerTrade = "maker"
)

//Exchange provides a generic wrapper for exchange services.
type Exchange interface {
	GetCandles(ctx context.Context, market *models.Market) ([]models.CandleStick, error)        // Gets the candle data from the exchange.
	GetMarketSummary(ctx context.Context, market *models.Market) (*models.MarketSummary, error) // Gets the current market summary.
	GetOrderBook(ctx context.Context, market *models.Market) (*models.OrderBook, error)         // Gets the order(ASK + BID) book of a market.

	BuyLimit(ctx context.Context, market *models.Market, amount float64, limit float64) (string, error)  // Performs a limit buy action.
	SellLimit(ctx context.Context, market *models.Market, amount float64, limit float64) (string, error) // Performs a limit sell action.
	BuyMarket(ctx context.Context, market *models.Market, amount float64) (string, error)                // Performs a market buy action.
	SellMarket(ctx context.Context, market *models.Market, amount float64) (string, error)               // Performs a market sell action.

	CalculateTradingFees(ctx context.Context, market *models.Market, amount float64, limit float64, orderType TradeType) float64 // Calculates the trading fees for an order on a specified market.
	CalculateWithdrawFees(ctx context.Context, market *models.Market, amount float64) float64                                    // Calculates the withdrawal fees on a specified market.

	GetBalance(ctx context.Context, symbol string) (*decimal.Decimal, error) // Gets the balance of the user of the specified currency.
	GetDepositAddress(coinTicker string) (string, bool)                      // Gets the deposit address for the specified coin on the exchange, if exists.

	Withdraw(ctx context.Context, destinationAddress string, coinTicker string, amount float64) error // Performs a withdraw operation from the exchange to a destination address.
}
