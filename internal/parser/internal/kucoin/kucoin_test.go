package kucoin_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/kucoin"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := kucoin.ParseSymbol("BTC-USDT")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDT", quote)

	base, quote, err = kucoin.ParseSymbol("BTC-USDC")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDC", quote)

	base, quote, err = kucoin.ParseSymbol("BCHSV-USDT")
	assert.NoError(t, err)
	assert.Equal(t, "BSV", base)
	assert.Equal(t, "USDT", quote)
}
