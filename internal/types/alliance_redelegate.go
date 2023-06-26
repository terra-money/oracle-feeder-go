package types

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

func DefaultAllianceRedelegateReq() *AllianceRedelegateReq {
	return &AllianceRedelegateReq{
		AllianceRedelegate: AllianceRedelegate{
			Redelegations: []Redelegation{},
		},
	}
}
