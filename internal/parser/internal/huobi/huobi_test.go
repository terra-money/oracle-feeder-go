package huobi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/huobi"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := huobi.ParseSymbol("btcusdt")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDT", quote)

	base, quote, err = huobi.ParseSymbol("btcusdc")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "USDC", quote)

	base, quote, err = huobi.ParseSymbol("etheuroc")
	assert.NoError(t, err)
	assert.Equal(t, "ETH", base)
	assert.Equal(t, "EUROC", quote)
}
