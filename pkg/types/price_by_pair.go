package types

// PriceByPair represents the USD price of a coin at a timestmap.
type PriceByPair struct {
	Base      string // Unified coin name, e.g., XBT is converted to BTC
	Quote     string // Unified coin name, e.g., XBT is converted to BTC
	Price     float64
	Timestamp uint64
}
