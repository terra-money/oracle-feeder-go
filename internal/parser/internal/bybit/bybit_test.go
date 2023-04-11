package bybit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/bybit"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := bybit.ParseSymbol("BTCUSDT")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDT", quote)

	base, quote, err = bybit.ParseSymbol("BTCUSD")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USD", quote)
}
