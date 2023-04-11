package parser

import (
	"fmt"
	"strings"

	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/binance"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/bitfinex"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/bitstamp"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/bybit"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/coingecko"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/huobi"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/kraken"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/kucoin"
)

// ParseSymbol parses exchange specific symbols to unified pairs.
//
// Each exchange has its own format for traiding symbols, for example,
// the two symbols, `BTCUSDT`of binance and `XBTUSDT` of Bitmex, both
// can be parsed to the same pair `BTC/USDT`.
func ParseSymbol(exhcange string, symbol string) (string, string, error) {
	switch strings.ToLower(exhcange) {
	case "binance":
		return binance.ParseSymbol(symbol)
	case "bitfinex":
		return bitfinex.ParseSymbol(symbol)
	case "bitstamp":
		return bitstamp.ParseSymbol(symbol)
	case "bybit":
		return bybit.ParseSymbol(symbol)
	case "coingecko":
		return coingecko.ParseSymbol(symbol)
	case "huobi":
		return huobi.ParseSymbol(symbol)
	case "kraken":
		return kraken.ParseSymbol(symbol)
	case "kucoin":
		return kucoin.ParseSymbol(symbol)
	default:
		return parseSymbolDefault(symbol)
	}
}

func parseSymbolDefault(symbol string) (string, string, error) {
	if strings.Contains(symbol, "/") {
		arr := strings.Split(symbol, "/")
		if len(arr) != 2 {
			return "", "", fmt.Errorf("cannot parse %s", symbol)
		}
		return arr[0], arr[1], nil
	} else if strings.Contains(symbol, "-") {
		arr := strings.Split(symbol, "-")
		if len(arr) != 2 {
			return "", "", fmt.Errorf("cannot parse %s", symbol)
		}
		return arr[0], arr[1], nil
	} else if strings.Contains(symbol, "_") {
		arr := strings.Split(symbol, "_")
		if len(arr) != 2 {
			return "", "", fmt.Errorf("cannot parse %s", symbol)
		}
		return arr[0], arr[1], nil
	} else {
		return "", "", fmt.Errorf("cannot parse %s", symbol)
	}
}
