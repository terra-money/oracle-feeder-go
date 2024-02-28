package coingecko

import (
	"fmt"
	"strings"
)

// symbol to base coin mapping
var COIN_GECKO_MAPPING = map[string]string{
	"bitcoin":                    "BTC",
	"ethereum":                   "ETH",
	"binancecoin":                "BNB",
	"tether":                     "USDT",
	"usd-coin":                   "USDC",
	"binance-usd":                "BUSD",
	"dai":                        "DAI",
	"okb":                        "OKB",
	"solana":                     "SOL",
	"cosmos":                     "ATOM",
	"terra-luna-2":               "LUNA",
	"terra-luna":                 "LUNC",
	"terrausd":                   "USTC",
	"injective-protocol":         "INJ",
	"eris-amplified-whale":       "AMPWHALE",
	"backbone-labs-staked-whale": "BWHALE",
	"secret":                     "SCRT",
	"juno-network":               "JUNO",
	"stargaze":                   "STARS",
	"akash-network":              "AKT",
	"lion-dao":                   "ROAR",   // Lion DAO's token
	"white-whale":                "WHALE",  // White Whale chain
	"switcheo":                   "SWTH",   // Carbon chain
	"stride-staked-luna":         "STLUNA", // Stride chain
	"stafi-staked-swth":          "rSWTH",  // stafi-staked-swth
	"osmosis":                    "OSMO",
}

func ParseSymbol(symbol string) (string, string, error) {
	symbol = strings.ToLower(symbol)
	if base, ok := COIN_GECKO_MAPPING[symbol]; ok {
		return base, "USD", nil
	} else {
		return "", "", fmt.Errorf("failed to parse CoinGecko %s", symbol)
	}
}
