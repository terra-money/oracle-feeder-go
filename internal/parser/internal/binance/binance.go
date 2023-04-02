package binance

import (
	"fmt"
	"strings"

	"github.com/terra-money/oracle-feeder-go/configs"
)

func ParseSymbol(symbol string) (string, string, error) {
	symbol = strings.ToUpper(symbol)
	for _, coin := range configs.FiatCoins {
		if strings.HasSuffix(symbol, coin) {
			base := strings.TrimSuffix(symbol, coin)
			return base, coin, nil
		}
	}
	for _, coin := range configs.StableCoins {
		if strings.HasSuffix(symbol, coin) {
			base := strings.TrimSuffix(symbol, coin)
			return base, coin, nil
		}
	}
	return "", "", fmt.Errorf("failed to parse Binance %s", symbol)
}
