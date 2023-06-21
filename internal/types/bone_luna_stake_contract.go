package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BoneLunaData struct {
	TotalUSTeak   string     `json:"total_usteak"`
	TotalNative   string     `json:"total_native"`
	ExchangeRate  sdk.Dec    `json:"exchange_rate"`
	UnlockedCoins []sdk.Coin `json:"unlocked_coins"`
}
