package coingecko_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/coingecko"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := coingecko.ParseSymbol("bitcoin")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USD", quote)

	base, quote, err = coingecko.ParseSymbol("ethereum")
	assert.NoError(t, err)
	assert.Equal(t, "ETH", base)
	assert.Equal(t, "USD", quote)

	base, quote, err = coingecko.ParseSymbol("tether")
	assert.NoError(t, err)
	assert.Equal(t, "USDT", base)
	assert.Equal(t, "USD", quote)

	base, quote, err = coingecko.ParseSymbol("usd-coin")
	assert.NoError(t, err)
	assert.Equal(t, "USDC", base)
	assert.Equal(t, "USD", quote)
}
