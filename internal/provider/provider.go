package provider

import (
	"fmt"
	"strings"

	"github.com/terra-money/oracle-feeder-go/configs"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal/osmosis"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
)

// Provider represents a source of prices.
//
// For example, a cryptocurrency exchange can be a provider.
type Provider interface {
	GetPrices() map[string]types.PriceByPair
}

func NewProvider(exchange string, config *configs.ProviderConfig, stopCh <-chan struct{}) (Provider, error) {
	switch strings.ToLower(exchange) {
	case "binance", "huobi", "kucoin", "bitfinex", "kraken", "okx", "coinbase":
		return internal.NewWebsocketProvider(exchange, config.Symbols, stopCh)
	case "coingecko", "bitstamp", "bittrex", "exchangerate", "fer", "frankfurter":
		return internal.NewRESTfulProvider(exchange, config.Symbols, config.Interval, config.Timeout, stopCh)
	case "osmosis":
		return osmosis.NewOsmosisProvider(config, stopCh)
	default:
		return nil, fmt.Errorf("unknown exchange %s", exchange)
	}
}
