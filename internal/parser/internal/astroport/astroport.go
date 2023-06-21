package astroport

import (
	"fmt"
	"strings"
)

// symbol to base coin mapping
var ASTRO_MAPPINGS = map[string]string{
	"ibc/B3504E092456BA618CC28AC671A71FB08C6CA0FD0BE7C8A5B5A3E2DD933CC9E4": "USDC",         // axlUSDC
	"ibc/CBF67A2BCF6CAE343FDF251E510C8E18C361FC02B23430C121116E0811835DEF": "USDT",         // axlUSDT
	"ibc/08095CEDEA29977C9DD0CE9A48329FDA622C183359D5F90CF04CC4FF80CBE431": "STLUNA",       // STLUNA
	"terra1ecgazyd0waaj3g7l9cmy5gulhxkps2gmxu9ghducvuypjq68mq2s5lvsct":     "AMPLUNA",      // ampLuna
	"terra17aj4ty4sz4yhgm08na8drc0v03v2jwr3waxcqrwhajj729zhl7zqnpc0ml":     "BACKBONELUNA", // backboneLuna
}

func ParseSymbol(symbol string) (base, quote string, err error) {
	symbolSplit := strings.Split(symbol, "-")
	base_symbol := symbolSplit[0]
	quote_symbol := symbolSplit[1]

	base, ok := ASTRO_MAPPINGS[base_symbol]
	if !ok {
		return base, quote, fmt.Errorf("failed to parse 'base_symbol' from 'ASTRO_MAPPINGS' %s", base_symbol)
	}

	quote, ok = ASTRO_MAPPINGS[quote_symbol]
	if !ok {
		return base, quote, fmt.Errorf("failed to parse 'quote_symbol' from 'ASTRO_MAPPINGS' %s", quote_symbol)
	}

	return base, quote, nil
}
