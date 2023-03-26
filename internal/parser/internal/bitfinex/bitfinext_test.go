package bitfinex_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/bitfinex"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := bitfinex.ParseSymbol("tBTCUSD")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USD", quote)

	base, quote, err = bitfinex.ParseSymbol("tBTCUST")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDT", quote)

	base, quote, err = bitfinex.ParseSymbol("t1INCH:USD")
	assert.NoError(t, err)
	assert.Equal(t, "1INCH", base)
	assert.Equal(t, "USD", quote)
}
