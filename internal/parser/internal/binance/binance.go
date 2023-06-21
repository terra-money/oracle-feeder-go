package binance

import (
	"fmt"
	"strings"

	"github.com/terra-money/oracle-feeder-go/config"
)

func ParseSymbol(symbol string) (string, string, error) {
	symbol = strings.ToUpper(symbol)
	for _, coin := range config.FiatCoins {
		if strings.HasSuffix(symbol, coin) {
			base := strings.TrimSuffix(symbol, coin)
			return base, coin, nil
		}
	}
	for _, coin := range config.StableCoins {
		if strings.HasSuffix(symbol, coin) {
			base := strings.TrimSuffix(symbol, coin)
			return base, coin, nil
		}
	}
	return "", "", fmt.Errorf("failed to parse Binance %s", symbol)
}
