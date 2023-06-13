package types

// PricesResponse represents the JSON response of all prices.
type PricesResponse struct {
	Timestamp string        `json:"created_at"` // RFC3339
	Prices    []PriceOfCoin `json:"prices,omitempty"`
}

// PriceResponse represents the JSON response for a specific price
type PriceResponse struct {
	Timestamp string      `json:"created_at"` // RFC3339
	Price     PriceOfCoin `json:"prices,omitempty"`
}
