package alliance_provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
	types "github.com/terra-money/oracle-feeder-go/internal/types"
	pkgtypes "github.com/terra-money/oracle-feeder-go/pkg/types"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	mintypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/terra-money/oracle-feeder-go/internal/provider"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal/carbon"
)

type allianceProtocolsInfo struct {
	internal.BaseGrpc
	provider.LSDProvider
	carbon.CarbonProvider
	config          *config.AllianceConfig
	providerManager *provider.ProviderManager
}

func NewAllianceProtocolsInfo(config *config.AllianceConfig, providerManager *provider.ProviderManager) *allianceProtocolsInfo {

	return &allianceProtocolsInfo{
		BaseGrpc:        *internal.NewBaseGrpc(),
		LSDProvider:     *provider.NewLSDProvider(),
		CarbonProvider:  *carbon.NewCarbonProvider(),
		config:          config,
		providerManager: providerManager,
	}
}

func (p *allianceProtocolsInfo) GetProtocolsInfo(ctx context.Context) (*pkgtypes.MsgUpdateChainsInfo, error) {
	protocolRes := types.DefaultAllianceProtocolRes()

	// Query the all prices at the beginning
	// to cache the data and avoid querying
	// the prices each time we query the protocols info.
	pricesRes := p.providerManager.GetPrices(ctx)

	// Given the default list of LSTSData this method
	// queries the blockchain for the rebase factor of the LSD.
	lstsData, err := p.queryRebaseFactors(p.config.LSTSData)
	if err != nil {
		fmt.Printf("queryRebaseFactors: %v \n", err)
		return nil, err
	}

	// Setup Luna price
	for _, price := range pricesRes.Prices {
		if strings.EqualFold(price.Denom, "LUNA") {
			luna, err := sdktypes.NewDecFromStr(strconv.FormatFloat(price.Price, 'f', -1, 64))
			if err != nil {
				fmt.Printf("parse price error for: %s %v \n", price.Denom, err)
				return nil, err
			}
			protocolRes.LunaPrice = luna
		}
	}

	// Iterate over all configured nodes in the config file,
	// create a grpcConnection to each node and query the required data.
	for _, grpcUrl := range p.config.GRPCUrls {
		grpcConn, err := p.BaseGrpc.Connection(grpcUrl, nil)
		if err != nil {
			return nil, err
		}
		defer grpcConn.Close()

		// Create the QueryClients for the necessary modules.
		nodeClient := tmservice.NewServiceClient(grpcConn)
		mintClient := mintypes.NewQueryClient(grpcConn)
		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		allianceClient := alliancetypes.NewQueryClient(grpcConn)

		// Request alliances from the origin chain.
		allianceRes, err := allianceClient.Alliances(ctx, &alliancetypes.QueryAlliancesRequest{})
		if err != nil {
			fmt.Printf("allianceRes: %v \n", err)
			return nil, err
		}
		if len(allianceRes.Alliances) == 0 {
			fmt.Printf("No alliances found on: %s \n", grpcUrl)
			continue
		}

		allianceParamsRes, err := allianceClient.Params(ctx, &alliancetypes.QueryParamsRequest{})
		if err != nil {
			fmt.Printf("allianceParamsRes: %v \n", err)
			return nil, err
		}

		nodeRes, err := nodeClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
		if err != nil {
			fmt.Printf("nodeRes: %v \n", err)
			return nil, err
		}

		stakingParamsRes, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
		if err != nil {
			fmt.Printf("stakingParamsRes: %v \n", err)
			return nil, err
		}

		var annualProvisionsRes *mintypes.QueryAnnualProvisionsResponse
		if strings.Contains(grpcUrl, "carbon") {
			annualProvisionsRes, err = p.CarbonProvider.GetAnnualProvisions(ctx)
		} else {
			annualProvisionsRes, err = mintClient.AnnualProvisions(ctx, &mintypes.QueryAnnualProvisionsRequest{})

		}

		if err != nil {
			fmt.Printf("annualProvisionsRes: %v \n", err)
			return nil, err
		}

		// Remove the "u" prefix from the bond denom and
		// search for the price of the native token in
		// the prices response.
		bondDenom := strings.Replace(stakingParamsRes.GetParams().BondDenom, "u", "", 1)
		var priceRes pkgtypes.PriceOfCoin

		for _, price := range pricesRes.Prices {
			if strings.EqualFold(price.Denom, bondDenom) {
				priceRes = price
			}
		}

		if priceRes.Denom == "" {
			return nil, fmt.Errorf("price not found for: %s", bondDenom)
		}

		price, err := sdktypes.NewDecFromStr(strconv.FormatFloat(priceRes.Price, 'f', -1, 64))
		if err != nil {
			fmt.Printf("parse price error for: %s %v \n", bondDenom, err)
			return nil, err
		}

		nativeToken := types.NewNativeToken(
			stakingParamsRes.GetParams().BondDenom,
			price,
			annualProvisionsRes.AnnualProvisions,
		)

		normalizedLunaAlliance := p.parseLunaAlliances(allianceParamsRes.Params, allianceRes.Alliances, lstsData)
		alliancesOnPhoenix := p.filterAlliancesOnPhoenix(nodeRes)

		protocolRes.ProtocolsInfo = append(protocolRes.ProtocolsInfo, types.NewProtocolInfo(
			nodeRes.DefaultNodeInfo.Network,
			nativeToken,
			normalizedLunaAlliance,
			alliancesOnPhoenix,
		))
	}
	res := pkgtypes.NewMsgUpdateChainsInfo(protocolRes)

	return &res, nil
}
func (p *allianceProtocolsInfo) queryRebaseFactors(configLST []config.LSTData) ([]config.LSTData, error) {
	for i, lst := range configLST {
		rebaseFactor, err := p.LSDProvider.QueryLSTRebaseFactor(lst.Symbol)
		if err != nil {
			fmt.Printf("queryRebaseFactors: %v \n", err)
			continue
		}
		configLST[i].RebaseFactor = *rebaseFactor
	}

	return configLST, nil

}

func (p *allianceProtocolsInfo) filterAlliancesOnPhoenix(nodeRes *tmservice.GetNodeInfoResponse) []types.BaseAlliance {
	baseAlliances := []types.BaseAlliance{}

	for _, allianceOnPhoenix := range p.config.LSTOnPhoenix {
		if allianceOnPhoenix.CounterpartyChainId == nodeRes.DefaultNodeInfo.Network {
			baseAlliances = append(baseAlliances, types.BaseAlliance{
				IBCDenom:     allianceOnPhoenix.IBCDenom,
				RebaseFactor: allianceOnPhoenix.RebaseFactor,
			})
		}
	}
	return baseAlliances
}

func (p *allianceProtocolsInfo) parseLunaAlliances(
	allianceParams alliancetypes.Params,
	alliances []alliancetypes.AllianceAsset,
	lstsData []config.LSTData,
) []types.LunaAlliance {
	lunaAlliances := []types.LunaAlliance{}

	for _, lstData := range lstsData {
		for _, alliance := range alliances {
			// When an alliance is whitelisted in lstData which
			// means that it is an alliance with an LSD of luna
			// so it must be included to the lunaAlliances.
			if strings.EqualFold(lstData.IBCDenom, alliance.Denom) {
				annualTakeRate := calculateAnnualizedTakeRate(allianceParams, alliance)
				normalizedRewardWeight := calculateNormalizedRewardWeight(allianceParams, alliances, alliance)

				lunaAlliances = append(lunaAlliances, types.NewLunaAlliance(
					alliance.Denom,
					normalizedRewardWeight,
					annualTakeRate,
					sdktypes.NewDecFromIntWithPrec(alliance.TotalTokens, 6),
					lstData.RebaseFactor,
				))
			}
		}
	}
	return lunaAlliances
}

func calculateAnnualizedTakeRate(
	params alliancetypes.Params,
	alliance alliancetypes.AllianceAsset,
) sdktypes.Dec {
	// When TakeRateClaimInterval is zero it means that users are not
	// receiving any rewards so annualized take rate is zero (right now).
	if params.TakeRateClaimInterval == 0 {
		return sdktypes.ZeroDec()
	}

	// Parse the take rate from any interval to takeRatePerYear ...
	takeRatePerYear := 365 / params.TakeRateClaimInterval.Hours() * 24

	// ... return the annualized take rate.
	return sdktypes.OneDec().
		Sub(sdktypes.
			OneDec().
			Sub(alliance.TakeRate).
			Power(uint64(takeRatePerYear)))
}

func calculateNormalizedRewardWeight(
	params alliancetypes.Params,
	alliances []alliancetypes.AllianceAsset,
	alliance alliancetypes.AllianceAsset,
) sdktypes.Dec {

	// We shouldd consider that reward weight
	// starts at one because it also takes in
	// consideration the OneDec.
	rewardsWeight := sdktypes.OneDec()
	for _, alliance := range alliances {
		// When an alliance is not initialized, it means that users are not
		// receiving rewards so NormalizedRewardWeight is zero (right now).
		if !alliance.IsInitialized {
			return sdktypes.ZeroDec()
		}
		// If an alliance is not initialized it means that
		// rewards are not distributed to that alliance so
		// it has a reward weight of zero.
		rewardsWeight = rewardsWeight.Add(alliance.RewardWeight)
	}

	return alliance.RewardWeight.Quo(rewardsWeight)
}
