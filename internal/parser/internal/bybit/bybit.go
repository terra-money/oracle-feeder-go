package bybit

import (
	"fmt"
	"strings"
)

func ParseSymbol(symbol string) (string, string, error) {
	var base string
	var quote string
	if strings.HasSuffix(symbol, "USDT") {
		quote = "USDT"
		base = strings.TrimSuffix(symbol, quote)
	} else if strings.HasSuffix(symbol, "USD") {
		quote = "USD"
		base = strings.TrimSuffix(symbol, quote)
	} else {
		return "", "", fmt.Errorf("cannot parse %s", symbol)
	}
	return base, quote, nil
}
