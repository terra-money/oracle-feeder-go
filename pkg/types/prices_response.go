package types

// PricesResponse represents the JSON response of all prices.
type PricesResponse struct {
	Timestap string        `json:"created_at"` // RFC3339
	Prices   []PriceOfCoin `json:"prices,omitempty"`
}
