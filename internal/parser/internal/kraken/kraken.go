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
	switch currency {
	case "XBT":
		return "BTC"
	default:
		return currency
	}
}
