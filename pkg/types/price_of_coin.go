package types

// PriceOfCoin represents the USD price of a coin at a timestmap.
type PriceOfCoin struct {
	Coin      string  `json:"denom"` // Unified coin name, e.g., XBT is converted to BTC
	Price     float64 `json:"price"`
	Timestamp uint64  `json:"-"`
}
