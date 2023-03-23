package huobi

import (
	"fmt"
	"strings"
)

var quotes = []string{
	"brl", "btc", "eth", "eur", "euroc", "gbp", "ht", "husd", "rub", "trx", "try", "tusd",
	"uah", "usdc", "usdd", "usdt", "ust", "ustc",
}

func ParseSymbol(symbol string) (string, string, error) {
	symbol = strings.ToLower(symbol)
	for _, coin := range quotes {
		if strings.HasSuffix(symbol, coin) {
			base := strings.TrimSuffix(symbol, coin)
			return strings.ToUpper(base), strings.ToUpper(coin), nil
		}
	}
	return "", "", fmt.Errorf("failed to parse Huobi %s", symbol)
}
