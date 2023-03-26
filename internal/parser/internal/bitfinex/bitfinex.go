package bitfinex

import (
	"fmt"
	"strings"
)

// https://api-pub.bitfinex.com/v2/conf/pub:map:currency:sym
var currencyMapping = map[string]string{
	"AAA":     "TESTAAA",
	"AIX":     "AI",
	"ALG":     "ALGO",
	"AMP":     "AMPL",
	"AMPF0":   "AMPLF0",
	"ATO":     "ATOM",
	"B21X":    "B21",
	"BBB":     "TESTBBB",
	"BCHABC":  "XEC",
	"BTCF0":   "BTC",
	"CNHT":    "CNHt",
	"DAT":     "DATA",
	"DOG":     "MDOGE",
	"DSH":     "DASH",
	"EDO":     "PNT",
	"ETH2P":   "ETH2Pending",
	"ETH2R":   "ETH2Rewards",
	"ETH2X":   "ETH2",
	"EUS":     "EURS",
	"EUT":     "EURt",
	"EUTF0":   "EURt",
	"FBT":     "FB",
	"GNT":     "GLM",
	"HIX":     "HI",
	"IDX":     "ID",
	"IOT":     "IOTA",
	"LBT":     "LBTC",
	"LES":     "LEO-EOS",
	"LET":     "LEO-ERC20",
	"LNX":     "LN-BTC",
	"MNA":     "MANA",
	"MXNT":    "MXNt",
	"OMN":     "OMNI",
	"PAS":     "PASS",
	"PBTCEOS": "pBTC-EOS",
	"PBTCETH": "PBTC-ETH",
	"PETHEOS": "pETH-EOS",
	"PLTCEOS": "PLTC-EOS",
	"PLTCETH": "PLTC-ETH",
	"QSH":     "QASH",
	"QTM":     "QTUM",
	"RBT":     "RBTC",
	"REP":     "REP2",
	"SNG":     "SNGLS",
	"STJ":     "STORJ",
	"SXX":     "SX",
	"TSD":     "TUSD",
	"UDC":     "USDC",
	"UST":     "USDt",
	"USTF0":   "USDt",
	"VSY":     "VSYS",
	"WBT":     "WBTC",
	"XAUT":    "XAUt",
	"XCH":     "XCHF",
	"YGG":     "MCS",
}

func ParseSymbol(symbol string) (string, string, error) {
	symbol = strings.TrimPrefix(symbol, "t")
	var base string
	var quote string
	if strings.Contains(symbol, ":") {
		arr := strings.Split(symbol, ":")
		if len(arr) != 2 {
			return "", "", fmt.Errorf("cannot parse %s", symbol)
		}
		quote = arr[1]
		base = arr[0]
	} else {
		quote = symbol[len(symbol)-3:]
		base = symbol[:len(symbol)-3]
	}
	return normalizeCurrency(base), normalizeCurrency(quote), nil
}

func normalizeCurrency(currency string) string {
	currency = strings.ToUpper(currency)
	item, ok := currencyMapping[currency]
	if ok {
		return strings.ToUpper(item)
	}
	return currency
}
