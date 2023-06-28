package alliance_provider

import (
	"context"

	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/pkg/types"

	"github.com/terra-money/oracle-feeder-go/internal/provider"
)

type allianceProvider struct {
	allianceProtocolsInfo      *allianceProtocolsInfo
	allianceValidatorsProvider *allianceValidatorsProvider
}

func NewAllianceProvider(config *config.AllianceConfig, providerManager *provider.ProviderManager) *allianceProvider {

	return &allianceProvider{
		allianceProtocolsInfo:      NewAllianceProtocolsInfo(config, providerManager),
		allianceValidatorsProvider: NewAllianceValidatorsProvider(config, providerManager),
	}
}

func (p *allianceProvider) GetProtocolsInfo(ctx context.Context) (*types.MsgUpdateChainsInfo, error) {
	return p.allianceProtocolsInfo.GetProtocolsInfo(ctx)
}

func (p *allianceProvider) GetAllianceRedelegateReq(ctx context.Context) (*types.MsgAllianceRedelegate, error) {
	return p.allianceValidatorsProvider.GetAllianceRedelegateReq(ctx)
}
