package bitstamp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terra-money/oracle-feeder-go/internal/parser/internal/bitstamp"
)

func TestParseSymbol(t *testing.T) {
	base, quote, err := bitstamp.ParseSymbol("ethusd")
	assert.NoError(t, err)
	assert.Equal(t, "ETH", base)
	assert.Equal(t, "USD", quote)

	base, quote, err = bitstamp.ParseSymbol("ethusdt")
	assert.NoError(t, err)
	assert.Equal(t, "ETH", base)
	assert.Equal(t, "USDT", quote)
}
