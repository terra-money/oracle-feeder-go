package kraken

import (
	"fmt"
	"strings"
)

func ParseSymbol(symbol string) (string, string, error) {
	symbol = strings.ToUpper(symbol)
	arr := strings.Split(symbol, "/")
	if len(arr) != 2 {
		return "", "", fmt.Errorf("failed to parse kraken %s", symbol)
	}
	return normalizeCurrency(arr[0]), normalizeCurrency(arr[1]), nil
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(currency)
	if len(currency) > 3 && (strings.HasPrefix(currency, "X") || strings.HasPrefix(currency, "Z")) {
		currency = currency[1:]
	}
	switch currency {
	case "XBT":
		return "BTC"
	case "XDG":
		return "DOGE"
	default:
		return currency
	}
}
