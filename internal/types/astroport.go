package types

type AstroportData struct {
	Path             Path     `json:"path"`
	Simulate         Simulate `json:"simulate"`
	AmountOut        int      `json:"amount_out"`
	TotalPriceImpact float64  `json:"total_price_impact"`
}

type Path struct {
	Route    []Route  `json:"route"`
	Tokens   []string `json:"tokens"`
	Illiquid bool     `json:"illiquid"`
}

type Route struct {
	ContractAddr string  `json:"contract_addr"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	Type         string  `json:"type"`
	PriceImpact  float64 `json:"price_impact"`
	Illiquid     bool    `json:"illiquid"`
}

type Simulate struct {
	SimulateSwapOperations SimulateSwapOperations `json:"simulate_swap_operations"`
}

type SimulateSwapOperations struct {
	OfferAmount string      `json:"offer_amount"`
	Operations  []Operation `json:"operations"`
}

type Operation struct {
	AstroSwap AstroSwap `json:"astro_swap"`
}

type AstroSwap struct {
	OfferAssetInfo OfferAssetInfo `json:"offer_asset_info"`
	AskAssetInfo   AskAssetInfo   `json:"ask_asset_info"`
}

type OfferAssetInfo struct {
	Token       *Token            `json:"token,omitempty"`
	NativeToken *AstroNativeToken `json:"native_token,omitempty"`
}

type AskAssetInfo struct {
	Token       *Token            `json:"token,omitempty"`
	NativeToken *AstroNativeToken `json:"native_token,omitempty"`
}

type Token struct {
	ContractAddr string `json:"contract_addr"`
}

type AstroNativeToken struct {
	Denom string `json:"denom"`
}
