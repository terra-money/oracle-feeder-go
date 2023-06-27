package types

import "github.com/cosmos/cosmos-sdk/types"

type Redelegation struct {
	SrcValidator string `json:"src_validator"`
	DstValidator string `json:"dst_validator"`
	Amount       string `json:"amount"`
}

func NewRedelegation(srcValidator string, dstValidator string, amount string) Redelegation {
	return Redelegation{
		SrcValidator: srcValidator,
		DstValidator: dstValidator,
		Amount:       amount,
	}
}

type AllianceRedelegate struct {
	Redelegations []Redelegation `json:"redelegations"`
}

type AllianceRedelegateReq struct {
	AllianceRedelegate AllianceRedelegate `json:"alliance_redelegate"`
}

func NewAllianceRedelegateReq(redelegations []Redelegation) *AllianceRedelegateReq {
	return &AllianceRedelegateReq{
		AllianceRedelegate: AllianceRedelegate{
			Redelegations: redelegations,
		},
	}
}

type ValWithAllianceTokensStake struct {
	ValidatorAddr string
	TotalStaked   types.DecCoin
}

func NewValWithAllianceTokensStake(validatorAddr string, totalStaked types.DecCoin) ValWithAllianceTokensStake {
	return ValWithAllianceTokensStake{
		ValidatorAddr: validatorAddr,
		TotalStaked:   totalStaked,
	}
}
