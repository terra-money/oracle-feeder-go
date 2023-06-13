package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	mintypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
	"github.com/terra-money/oracle-feeder-go/config"
	types "github.com/terra-money/oracle-feeder-go/internal/types"
	pricetypes "github.com/terra-money/oracle-feeder-go/pkg/types"
	"google.golang.org/grpc"
)

type allianceProvider struct {
	config          *config.AllianceConfig
	providerManager *ProviderManager
}

func NewAllianceProvider(config *config.AllianceConfig, providerManager *ProviderManager) *allianceProvider {

	return &allianceProvider{
		config:          config,
		providerManager: providerManager,
	}
}

func (p *allianceProvider) getRPCConnection(nodeUrl string, interfaceRegistry sdk.InterfaceRegistry) (*grpc.ClientConn, error) {
	return grpc.Dial(
		nodeUrl,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(interfaceRegistry).GRPCCodec())),
	)
}

func (p *allianceProvider) GetProtocolsInfo(ctx context.Context) (protocolsInfo []types.ProtocolInfo, err error) {
	// Query the all prices at the beggining
	// to cache the data and avoid querying
	// the prices each time we query the protocols info.
	pricesRes := p.providerManager.GetPrices(ctx)

	// Iterate over all configured nodes in the config file,
	// create a grpcConnection to each node and query the required data.
	for _, grpcUrl := range p.config.GRPCUrls {
		grpcConn, err := p.getRPCConnection(grpcUrl, nil)
		if err != nil {
			return nil, err
		}
		defer grpcConn.Close()

		// Create the QueryClients for the necessary modules.
		nodeClient := tmservice.NewServiceClient(grpcConn)
		mintClient := mintypes.NewQueryClient(grpcConn)
		stakingClient := stakingtypes.NewQueryClient(grpcConn)

		// Request alliances from the origin chain.
		alliances, err := queryAlliances(ctx, grpcConn)
		if err != nil {
			continue
		}

		nodeRes, err := nodeClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
		if err != nil {
			fmt.Println("nodeRes:", err)
			continue
		}

		stakingParamsRes, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
		if err != nil {
			fmt.Println("stakingParamsRes:", err)
			continue
		}

		inflationRes, err := mintClient.Inflation(ctx, &mintypes.QueryInflationRequest{})
		if err != nil {
			fmt.Println("inflationRes:", err)
			continue
		}

		// Remove the "u" prefix from the bond denom and
		// search for the price of the native token in
		// the prices response.
		bondDenom := strings.Replace(stakingParamsRes.GetParams().BondDenom, "u", "", 1)
		var priceRes pricetypes.PriceOfCoin

		for _, price := range pricesRes.Prices {
			if strings.EqualFold(price.Denom, bondDenom) {
				priceRes = price
			}
		}

		nativeToken := types.NewNativeToken(
			stakingParamsRes.GetParams().BondDenom,
			priceRes.Price,
			inflationRes.Inflation,
		)

		normalizedLunaAlliance := normalizeLunaAlliance(alliances)
		lstsData := parseLtsData(alliances, p.config.LSTSData)

		protocolsInfo = append(protocolsInfo, types.NewProtocolInfo(
			nodeRes.DefaultNodeInfo.Network,
			nativeToken,
			normalizedLunaAlliance,
			lstsData,
		))
	}

	return protocolsInfo, err
}

// Query all alliances from the origin chain
// and filter out the ones that are not
// related to Luna.
func queryAlliances(ctx context.Context, grpcConn *grpc.ClientConn) (alliances []alliancetypes.AllianceAsset, err error) {
	ibcClient := ibctypes.NewQueryClient(grpcConn)
	allianceClient := alliancetypes.NewQueryClient(grpcConn)
	res, err := allianceClient.Alliances(ctx, &alliancetypes.QueryAlliancesRequest{})
	if err != nil {
		return nil, err
	}

	if len(res.Alliances) == 0 {
		return nil, fmt.Errorf("no alliances found")
	}

	for _, alliance := range res.Alliances {
		ibcDenomRes, err := ibcClient.DenomTrace(ctx, &ibctypes.QueryDenomTraceRequest{
			Hash: strings.Split(alliance.Denom, "ibc/")[1],
		})
		if err != nil {
			fmt.Println("ibcDenomRes:", err)
			continue
		}

		if !strings.HasPrefix(ibcDenomRes.DenomTrace.BaseDenom, "cw20:terra") &&
			!strings.Contains(ibcDenomRes.DenomTrace.BaseDenom, "luna") {
			continue
		}

		alliances = append(alliances, alliance)
	}

	return alliances, nil
}

// Since there are multiple LSD's in the ecosystem,
// it's necessary to normalize the data, so we can
// get the average of all the LSD's.
func normalizeLunaAlliance(alliances []alliancetypes.AllianceAsset) types.NormalizedLunaAlliance {
	normalizedAlliances := types.DefaultNormalizedLunaAlliance()

	for _, alliance := range alliances {
		normalizedAlliances.RewardsWeight = normalizedAlliances.RewardsWeight.Add(alliance.RewardWeight)
		normalizedAlliances.LSDStaked = normalizedAlliances.LSDStaked.Add(alliance.TotalTokens)
		normalizedAlliances.AnnualTakeRate = normalizedAlliances.AnnualTakeRate.Add(alliance.TakeRate)
	}

	normalizedAlliances.RewardsWeight = normalizedAlliances.RewardsWeight.QuoInt64(int64(len(alliances)))
	normalizedAlliances.LSDStaked = normalizedAlliances.LSDStaked.Quo(sdktypes.NewInt(int64(len(alliances))))
	normalizedAlliances.AnnualTakeRate = normalizedAlliances.AnnualTakeRate.QuoInt64(int64(len(alliances)))

	return normalizedAlliances
}

// Filter all alliances to match the ones in
// the config file to return the LTS data.
func parseLtsData(alliances []alliancetypes.AllianceAsset, lstsData []config.LSTData) (lsts []config.LSTData) {
	for _, alliance := range alliances {
		for _, lstData := range lstsData {
			if strings.EqualFold(lstData.IBCDenom, alliance.Denom) {
				lsts = append(lsts, lstData)
			}
		}
	}

	return lsts
}
