package types

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProtocolInfo struct {
	ChainId       string         `json:"chain_id,omitempty"`
	NativeToken   NativeToken    `json:"native_token,omitempty"`
	LunaAlliances []LunaAlliance `json:"luna_alliances,omitempty"`
}

func NewProtocolInfo(
	chainId string,
	nativeToken NativeToken,
	lunaAlliances []LunaAlliance,
) ProtocolInfo {
	return ProtocolInfo{
		ChainId:       chainId,
		NativeToken:   nativeToken,
		LunaAlliances: lunaAlliances,
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
	IBCDenom               string   `json:"ibc_denom,omitempty"`
	NormalizedRewardWeight sdk.Dec  `json:"normalized_reward_weight,omitempty"`
	AnnualTakeRate         sdk.Dec  `json:"annual_take_rate,omitempty"`
	TotalStaked            math.Int `json:"total_staked,omitempty"`
	RebaseFactor           sdk.Dec  `json:"rebase_factor,omitempty"`
}

func NewLunaAlliance(
	ibcDenom string,
	normalizedRewardWeight,
	annualTakeRate sdk.Dec,
	totalStaked math.Int,
	rebaseFactor sdk.Dec,
) LunaAlliance {
	return LunaAlliance{
		IBCDenom:               ibcDenom,
		NormalizedRewardWeight: normalizedRewardWeight,
		AnnualTakeRate:         annualTakeRate,
		TotalStaked:            totalStaked,
		RebaseFactor:           rebaseFactor,
	}
}

func DefaultNormalizedLunaAlliance() LunaAlliance {
	return LunaAlliance{
		IBCDenom:               "",
		NormalizedRewardWeight: sdk.ZeroDec(),
		AnnualTakeRate:         sdk.ZeroDec(),
		TotalStaked:            sdk.ZeroInt(),
		RebaseFactor:           sdk.ZeroDec(),
	}
}
