package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ErisData struct {
	TotalUSTake   string     `json:"total_ustake"`
	TotalULuna    string     `json:"total_uluna"`
	ExchangeRate  sdk.Dec    `json:"exchange_rate"`
	UnlockedCoins []sdk.Coin `json:"unlocked_coins"`
	Unbonding     string     `json:"unbonding"`
	Available     string     `json:"available"`
	TVLULuna      string     `json:"tvl_uluna"`
}
