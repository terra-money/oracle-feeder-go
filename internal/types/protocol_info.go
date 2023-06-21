package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AllianceProtocolRes struct {
	LunaPrice     sdk.Dec        `json:"luna_price,omitempty"`
	ProtocolsInfo []ProtocolInfo `json:"protocols_info,omitempty"`
}

func DefaultAllianceProtocolRes() AllianceProtocolRes {
	return AllianceProtocolRes{
		LunaPrice:     sdk.ZeroDec(),
		ProtocolsInfo: []ProtocolInfo{},
	}
}

type ProtocolInfo struct {
	ChainId                 string         `json:"chain_id,omitempty"`
	NativeToken             NativeToken    `json:"native_token,omitempty"`
	LunaAlliances           []LunaAlliance `json:"luna_alliances"`
	ChainAlliancesOnPhoenix []BaseAlliance `json:"chain_alliances_on_phoenix"`
}

func NewProtocolInfo(
	chainId string,
	nativeToken NativeToken,
	lunaAlliances []LunaAlliance,
	chainAlliancesOnPhoenix []BaseAlliance,
) ProtocolInfo {
	return ProtocolInfo{
		ChainId:                 chainId,
		NativeToken:             nativeToken,
		LunaAlliances:           lunaAlliances,
		ChainAlliancesOnPhoenix: chainAlliancesOnPhoenix,
	}
}

type NativeToken struct {
	Denom            string  `json:"denom,omitempty"`
	TokenPrice       sdk.Dec `json:"token_price,omitempty"`
	AnnualProvisions sdk.Dec `json:"annual_provisions,omitempty"`
}

func NewNativeToken(
	denom string,
	tokenPrice sdk.Dec,
	annualProvisions sdk.Dec,
) NativeToken {
	return NativeToken{
		Denom:            denom,
		TokenPrice:       tokenPrice,
		AnnualProvisions: annualProvisions,
	}
}

type BaseAlliance struct {
	IBCDenom     string  `json:"ibc_denom,omitempty"`
	RebaseFactor sdk.Dec `json:"rebase_factor,omitempty"`
}

func NewBaseAlliance(
	ibcDenom string,
	rebaseFactor sdk.Dec,
) BaseAlliance {
	return BaseAlliance{
		IBCDenom:     ibcDenom,
		RebaseFactor: rebaseFactor,
	}
}

type LunaAlliance struct {
	BaseAlliance
	NormalizedRewardWeight sdk.Dec `json:"normalized_reward_weight,omitempty"`
	AnnualTakeRate         sdk.Dec `json:"annual_take_rate,omitempty"`
	TotalLSDStaked         sdk.Dec `json:"total_lsd_staked,omitempty"`
}

func NewLunaAlliance(
	ibcDenom string,
	normalizedRewardWeight,
	annualTakeRate sdk.Dec,
	totalLSDStaked sdk.Dec,
	rebaseFactor sdk.Dec,
) LunaAlliance {
	baseAlliance := BaseAlliance{
		IBCDenom:     ibcDenom,
		RebaseFactor: rebaseFactor,
	}
	return LunaAlliance{
		BaseAlliance:           baseAlliance,
		NormalizedRewardWeight: normalizedRewardWeight,
		AnnualTakeRate:         annualTakeRate,
		TotalLSDStaked:         totalLSDStaked,
	}
}
