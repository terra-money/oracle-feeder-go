package binance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/binance"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := binance.ParseSymbol("BTCUSDT")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDT", quote)

	base, quote, err = binance.ParseSymbol("BTCUSDC")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDC", quote)

	base, quote, err = binance.ParseSymbol("BTCUSD")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USD", quote)
}
