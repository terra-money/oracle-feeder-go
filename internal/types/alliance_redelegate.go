package types

import "github.com/cosmos/cosmos-sdk/types"

type Delegation struct {
	Validator string `json:"validator"`
	Amount    string `json:"amount"`
}

func NewDelegation(validator string, amount string) Delegation {
	return Delegation{
		Validator: validator,
		Amount:    amount,
	}
}

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
