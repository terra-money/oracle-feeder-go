package bitstamp

import (
	"strings"
)

func ParseSymbol(symbol string) (string, string, error) {
	var base string
	var quote string
	if strings.HasSuffix(symbol, "usdc") || strings.HasSuffix(symbol, "usdt") {
		quote = symbol[len(symbol)-4:]
		base = symbol[:len(symbol)-4]
	} else {
		quote = symbol[len(symbol)-3:]
		base = symbol[:len(symbol)-3]
	}
	return strings.ToUpper(base), strings.ToUpper(quote), nil
}
