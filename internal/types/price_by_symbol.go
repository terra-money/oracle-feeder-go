package types

// PriceBySymbol represents the price of a trading symbol at a timestmap.
type PriceBySymbol struct {
	Exchange  string
	Symbol    string // Exchange-specific trading symbol, e.g., XBTUSD from bitmex
	Base      string
	Quote     string
	Price     float64
	Timestamp uint64
}
