package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProtocolInfo struct {
	ChainId        string         `json:"chain_id,omitempty"`
	NativeToken    NativeToken    `json:"native_token,omitempty"`
	LunaAlliances  []LunaAlliance `json:"luna_alliances,omitempty"`
	LastUpdateTime time.Time      `json:"last_update_time,omitempty"`
}

func NewProtocolInfo(
	chainId string,
	nativeToken NativeToken,
	lunaAlliances []LunaAlliance,
) ProtocolInfo {
	return ProtocolInfo{
		ChainId:        chainId,
		NativeToken:    nativeToken,
		LunaAlliances:  lunaAlliances,
		LastUpdateTime: time.Now(),
	}
}

type NativeToken struct {
	Denom           string  `json:"denom,omitempty"`
	TokenPrice      sdk.Dec `json:"token_price,omitempty"`
	AnnualInflation sdk.Dec `json:"annual_inflation,omitempty"`
}

func NewNativeToken(
	denom string,
	tokenPrice sdk.Dec,
	annualInflation sdk.Dec,
) NativeToken {
	return NativeToken{
		Denom:           denom,
		TokenPrice:      tokenPrice,
		AnnualInflation: annualInflation,
	}
}

type LunaAlliance struct {
	IBCDenom       string  `json:"ibc_denom,omitempty"`
	RewardWeight   sdk.Dec `json:"reward_weight,omitempty"`
	AnnualTakeRate sdk.Dec `json:"annual_take_rate,omitempty"`
	TotalStaked    sdk.Int `json:"total_staked,omitempty"`
	RebaseFactor   sdk.Dec `json:"rebase_factor,omitempty"`
}

func NewLunaAlliance(
	ibcDenom string,
	rewardWeight,
	annualTakeRate sdk.Dec,
	totalStaked sdk.Int,
	rebaseFactor sdk.Dec,
) LunaAlliance {
	return LunaAlliance{
		IBCDenom:       ibcDenom,
		RewardWeight:   rewardWeight,
		AnnualTakeRate: annualTakeRate,
		TotalStaked:    totalStaked,
		RebaseFactor:   rebaseFactor,
	}
}

func DefaultNormalizedLunaAlliance() LunaAlliance {
	return LunaAlliance{
		IBCDenom:       "",
		RewardWeight:   sdk.ZeroDec(),
		AnnualTakeRate: sdk.ZeroDec(),
		TotalStaked:    sdk.ZeroInt(),
		RebaseFactor:   sdk.ZeroDec(),
	}
}
