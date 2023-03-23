package configs

var FiatCoins = map[string]bool{
	"AUD":  true,
	"BIDR": true,
	"BRL":  true,
	"CHF":  true,
	"CNH":  true,
	"CNY":  true,
	"EUR":  true,
	"GBP":  true,
	"IDRT": true,
	"INR":  true,
	"JPY":  true,
	"KRW":  true,
	"MXN":  true,
	"NGN":  true,
	"NZD":  true,
	"PLN":  true,
	"RON":  true,
	"RUB":  true,
	"SEK":  true,
	"TRY":  true,
	"UAH":  true,
	"USD":  true,
	"ZAR":  true,
}

// https://coinmarketcap.com/view/stablecoin/
// https://www.stablecoinswar.com/
var StableCoins = map[string]bool{
	"USDT":  true,
	"USDC":  true,
	"BUSD":  true,
	"DAI":   true,
	"TUSD":  true,
	"USDP":  true,
	"USDD":  true,
	"GUSD":  true,
	"FEI":   true,
	"USTC":  true,
	"TRIBE": true,
	"FRAX":  true,
	"USDJ":  true,
	"LUSD":  true,
	"EURS":  true,
	"USDX":  true,
}
