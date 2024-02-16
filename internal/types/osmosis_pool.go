package types

// PriceBySymbol represents the price of a trading symbol at a timestamp.
type OsmosisConcentratedLiquidityPool struct {
	Type                 string `json:"@type"`
	Address              string `json:"address"`
	IncentivesAddress    string `json:"incentives_address"`
	SpreadRewardsAddress string `json:"spread_rewards_address"`
	ID                   string `json:"id"`
	CurrentTickLiquidity string `json:"current_tick_liquidity"`
	Token0               string `json:"token0"`
	Token1               string `json:"token1"`
	CurrentSqrtPrice     string `json:"current_sqrt_price"`
	CurrentTick          string `json:"current_tick"`
	TickSpacing          string `json:"tick_spacing"`
	ExponentAtPriceOne   string `json:"exponent_at_price_one"`
	SpreadFactor         string `json:"spread_factor"`
	LastLiquidityUpdate  string `json:"last_liquidity_update"`
}

type OsmosisPoolResponse struct {
	Pool OsmosisConcentratedLiquidityPool `json:"pool"`
}

type OsmoToken struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type PoolAsset struct {
	Token  OsmoToken `json:"token"`
	Weight string    `json:"weight"`
}

type PoolParams struct {
	SwapFee                  string      `json:"swap_fee"`
	ExitFee                  string      `json:"exit_fee"`
	SmoothWeightChangeParams interface{} `json:"smooth_weight_change_params"`
}

type TotalShares struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type OsmosisGammPool struct {
	Type               string      `json:"@type"`
	Address            string      `json:"address"`
	ID                 string      `json:"id"`
	PoolParams         PoolParams  `json:"pool_params"`
	FuturePoolGovernor string      `json:"future_pool_governor"`
	TotalShares        TotalShares `json:"total_shares"`
	PoolAssets         []PoolAsset `json:"pool_assets"`
	TotalWeight        string      `json:"total_weight"`
}

type OsmosisGammPoolResponse struct {
	Pool OsmosisGammPool `json:"pool"`
}

type GenericPoolResponse struct {
	Pool GenericType `json:"pool"`
}

type GenericType struct {
	Type string `json:"@type"`
}
