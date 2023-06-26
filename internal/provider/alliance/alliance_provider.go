package alliance_provider

import (
	"context"

	"github.com/terra-money/oracle-feeder-go/config"
	types "github.com/terra-money/oracle-feeder-go/internal/types"

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

func (p *allianceProvider) GetProtocolsInfo(ctx context.Context) (*types.AllianceProtocolRes, error) {
	return p.allianceProtocolsInfo.GetProtocolsInfo(ctx)
}

func (p *allianceProvider) GetAllianceRedelegateReq(ctx context.Context) (*types.AllianceRedelegateReq, error) {
	return p.allianceValidatorsProvider.GetAllianceRedelegateReq(ctx)
}
