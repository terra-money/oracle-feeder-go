package types

// PriceOfCoin represents the USD price of a coin at a timestamp.
type PriceOfCoin struct {
	Denom     string  `json:"denom"` // Unified denom name, e.g., XBT is converted to BTC
	Price     float64 `json:"price"`
	Timestamp uint64  `json:"-"`
}
