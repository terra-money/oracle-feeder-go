package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/oracle-feeder-go/config"
)

type ProtocolInfo struct {
	ChainId                string                 `json:"chain_id,omitempty"`
	NativeToken            NativeToken            `json:"native_token,omitempty"`
	NormalizedLunaAlliance NormalizedLunaAlliance `json:"luna_alliance_data,omitempty"`
	LSTSData               []config.LSTData       `json:"lst_data,omitempty"`
}

func NewProtocolInfo(
	chainId string,
	nativeToken NativeToken,
	normalizedLunaAlliance NormalizedLunaAlliance,
	lstsData []config.LSTData,
) ProtocolInfo {
	return ProtocolInfo{
		ChainId:                chainId,
		NativeToken:            nativeToken,
		NormalizedLunaAlliance: normalizedLunaAlliance,
		LSTSData:               lstsData,
	}
}

type NativeToken struct {
	Denom           string  `json:"denom,omitempty"`
	TokenPrice      float64 `json:"rebase_factor,omitempty"`
	AnnualInflation sdk.Dec `json:"annual_inflation,omitempty"`
}

func NewNativeToken(
	denom string,
	tokenPrice float64,
	annualInflation sdk.Dec,
) NativeToken {
	return NativeToken{
		Denom:           denom,
		TokenPrice:      tokenPrice,
		AnnualInflation: annualInflation,
	}
}

type NormalizedLunaAlliance struct {
	RewardsWeight  sdk.Dec `json:"rewards_weight,omitempty"`
	AnnualTakeRate sdk.Dec `json:"annual_take_rate,omitempty"`
	LSDStaked      sdk.Int `json:"lsd_staked,omitempty"`
}

func NewNormalizedLunaAlliance(
	rewardsWeight,
	annualTakeRate sdk.Dec,
	lsdStaked sdk.Int,
) NormalizedLunaAlliance {
	return NormalizedLunaAlliance{
		RewardsWeight:  rewardsWeight,
		AnnualTakeRate: annualTakeRate,
		LSDStaked:      lsdStaked,
	}
}

func DefaultNormalizedLunaAlliance() NormalizedLunaAlliance {
	return NormalizedLunaAlliance{
		RewardsWeight:  sdk.ZeroDec(),
		AnnualTakeRate: sdk.ZeroDec(),
		LSDStaked:      sdk.ZeroInt(),
	}
}
