package bitfinex

import (
	"fmt"
	"strings"

	"github.com/terra-money/oracle-feeder-go/configs"
)

func ParseSymbol(symbol string) (string, string, error) {
	if strings.HasPrefix(symbol, "t") {
		symbol = symbol[1:]
	}
	if strings.Contains(symbol, ":") {
		arr := strings.Split(symbol, ":")
		if len(arr) != 2 {
			return "", "", fmt.Errorf("cannot parse %s", symbol)
		}
		return normalizeCurrency(arr[0]), normalizeCurrency(arr[1]), nil
	}
	symbol = strings.ToUpper(symbol)
	for coin := range configs.FiatCoins {
		if strings.HasSuffix(symbol, coin) {
			base := strings.TrimSuffix(symbol, coin)
			return normalizeCurrency(base), normalizeCurrency(coin), nil
		}
	}
	for coin := range configs.StableCoins {
		if strings.HasSuffix(symbol, coin) {
			base := strings.TrimSuffix(symbol, coin)
			return normalizeCurrency(base), normalizeCurrency(coin), nil
		}
	}
	return "", "", fmt.Errorf("failed to parse bitfinex %s", symbol)
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(currency)
	switch currency {
	case "UST":
		return "USDT"
	default:
		return currency
	}
}
