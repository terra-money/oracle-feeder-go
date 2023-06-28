package types

import (
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

type MsgUpdateChainsInfo struct {
	UpdateChainsInfo UpdateChainsInfo `json:"update_chains_info"`
}

type UpdateChainsInfo struct {
	ChainsInfo types.AllianceProtocolRes `json:"chains_info"`
}

func NewMsgUpdateChainsInfo(data types.AllianceProtocolRes) MsgUpdateChainsInfo {
	return MsgUpdateChainsInfo{
		UpdateChainsInfo: UpdateChainsInfo{
			ChainsInfo: data,
		},
	}
}

type AllianceRedelegate struct {
	Redelegations []types.Redelegation `json:"redelegations"`
}
type MsgAllianceRedelegate struct {
	AllianceRedelegate AllianceRedelegate `json:"alliance_redelegate"`
}

func NewMsgAllianceRedelegate(redelegations []types.Redelegation) MsgAllianceRedelegate {
	return MsgAllianceRedelegate{
		AllianceRedelegate: AllianceRedelegate{
			Redelegations: redelegations,
		},
	}
}
