package types

import (
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

type UpdateDataMsg struct {
	UpdateData []types.ProtocolInfo `json:"update_data"`
}

func NewUpdateDataMsg(data []types.ProtocolInfo) UpdateDataMsg {
	return UpdateDataMsg{
		UpdateData: data,
	}
}
