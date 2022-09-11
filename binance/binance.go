package binance

import (
	"github.com/adshao/go-binance/v2"
	"github.com/anvh2/trading-boy/logger"
)

// BinanceWrapper represents the wrapper for the Binance exchange.
type BinanceWrapper struct {
	logger           *logger.Logger
	api              *binance.Client
	depositAddresses map[string]string
}

func NewBinanceWrapper(logger *logger.Logger, publicKey string, secretKey string, depositAddresses map[string]string) Exchange {
	client := binance.NewClient(publicKey, secretKey)
	return &BinanceWrapper{
		logger:           logger,
		api:              client,
		depositAddresses: depositAddresses,
	}
}
func (bw *BinanceWrapper) GetDepositAddress(coinTicker string) (string, bool) {
	addr, exists := bw.depositAddresses[coinTicker]
	return addr, exists
}

func (bw *BinanceWrapper) ListDepositAddress() map[string]string {
	return bw.depositAddresses
}
