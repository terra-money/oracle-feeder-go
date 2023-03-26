package kraken_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/kraken"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := kraken.ParseSymbol("PAXG/USD")
	assert.NoError(t, err)
	assert.Equal(t, "PAXG", base)
	assert.Equal(t, "USD", quote)

	base, quote, err = kraken.ParseSymbol("XBT/USDC")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDC", quote)

	base, quote, err = kraken.ParseSymbol("XDG/USD")
	assert.NoError(t, err)
	assert.Equal(t, "DOGE", base)
	assert.Equal(t, "USD", quote)
}
