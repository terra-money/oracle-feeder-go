package okx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/okx"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := okx.ParseSymbol("1INCH-USDC")
	assert.NoError(t, err)
	assert.Equal(t, "1INCH", base)
	assert.Equal(t, "USDC", quote)

	base, quote, err = okx.ParseSymbol("BTC-DAI")
	assert.NoError(t, err)
	assert.Equal(t, "BTC", base)
	assert.Equal(t, "DAI", quote)
}
