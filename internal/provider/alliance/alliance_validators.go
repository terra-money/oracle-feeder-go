package alliance_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
	types "github.com/terra-money/oracle-feeder-go/internal/types"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type allianceValidatorsProvider struct {
	internal.BaseGrpc
	nodeGrpcUrl     string
	stationApiUrl   string
	config          *config.AllianceConfig
	providerManager *provider.ProviderManager
}

func NewAllianceValidatorsProvider(config *config.AllianceConfig, providerManager *provider.ProviderManager) *allianceValidatorsProvider {

	var nodeGrpcUrl string
	if nodeGrpcUrl = os.Getenv("NODE_GRPC_URL"); len(nodeGrpcUrl) == 0 {
		panic("NODE_GRPC_URL env variable is not set!")
	}

	var stationApiUrl string
	if stationApiUrl = os.Getenv("STATION_API"); len(stationApiUrl) == 0 {
		panic("STATION_API env variable is not set!")
	}
	return &allianceValidatorsProvider{
		BaseGrpc:        *internal.NewBaseGrpc(),
		config:          config,
		nodeGrpcUrl:     nodeGrpcUrl,
		stationApiUrl:   stationApiUrl,
		providerManager: providerManager,
	}
}

// Query Terra GRPC, station API and return the list of alliance
// redelegations that can be submitted to Alliance Hub applying
// the following rules:
// (1) - be part of the active validator set,
// (2) - to do not be jailed,
// (3) - commission rate to be lower than 19%,
// (4) - Participate in the latest 3 gov proposals,
// (5) - have been in the active validator set 100 000 blocks before the current one, (1 week approx)
func (p *allianceValidatorsProvider) GetAllianceRedelegateReq(ctx context.Context) (*types.AllianceRedelegateReq, error) {
	allianceRebalanceValsRes := types.DefaultAllianceRedelegateReq()

	valsRes, seniorValidatorsRes, proposalsVotesRes, err := p.queryValidatorsData(ctx)
	if err != nil {
		return allianceRebalanceValsRes, err
	}

	var vals []stakingtypes.Validator
	// Apply the previous rules to filter the list
	// of all blockchain validators.
	for _, val := range valsRes.Validators {
		// (1) if status is not bonded continue (again in case the api has a bug with the query)
		if val.GetStatus() != stakingtypes.Bonded {
			continue
		}

		// (2) if jailed continue
		if val.IsJailed() {
			continue
		}

		// (3) if commission is grater than 19% continue
		if val.Commission.CommissionRates.Rate.GT(sdktypes.MustNewDecFromStr("0.19")) {
			continue
		}

		// (4) has voted in the last 3 proposals
		if !atLeastThreeOccurrences(proposalsVotesRes, val.OperatorAddress) {
			continue
		}

		// (5) has been in the active validator set 100 000 blocks before the current one
		for _, seniorValidator := range seniorValidatorsRes.Validators {
			if val.OperatorAddress != seniorValidator.Address {
				continue
			}
		}

		vals = append(vals, val)

		allianceRebalanceValsRes.AllianceRedelegate.Redelegations = append(
			allianceRebalanceValsRes.AllianceRedelegate.Redelegations,
			types.NewRedelegation(
				val.OperatorAddress,
				"",
				"",
			),
		)
	}

	return allianceRebalanceValsRes, nil
}

func (p *allianceValidatorsProvider) queryValidatorsData(ctx context.Context) (*stakingtypes.QueryValidatorsResponse, *tmservice.GetValidatorSetByHeightResponse, []types.StationVote, error) {
	grpcConn, err := p.BaseGrpc.Connection(p.nodeGrpcUrl, nil)
	if err != nil {
		fmt.Printf("grpcConn: %v \n", err)
		return nil, nil, nil, err
	}
	defer grpcConn.Close()

	nodeClient := tmservice.NewServiceClient(grpcConn)
	govClient := govtypes.NewQueryClient(grpcConn)
	stakingClient := stakingtypes.NewQueryClient(grpcConn)

	valsRes, err := stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded, // (1) query only status bonded
		Pagination: &query.PageRequest{
			Limit: 150,
		},
	})
	if err != nil {
		fmt.Printf("valsRes: %v \n", err)
		return nil, nil, nil, err
	}

	govPropsRes, err := govClient.Proposals(ctx, &govtypes.QueryProposalsRequest{
		Pagination: &query.PageRequest{
			Limit:   3,
			Reverse: true,
		},
	})
	if err != nil {
		fmt.Printf("govPropsRes: %v \n", err)
		return nil, nil, nil, err
	}

	latestHeightRes, err := nodeClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		fmt.Printf("latestHeightRes: %v \n", err)
		return nil, nil, nil, err
	}

	seniorValidatorsRes, err := nodeClient.GetValidatorSetByHeight(ctx, &tmservice.GetValidatorSetByHeightRequest{
		Height: latestHeightRes.SdkBlock.Header.Height - 100_000,
	})
	if err != nil {
		fmt.Printf("seniorValidatorsRes: %v \n", err)
		return nil, nil, nil, err
	}

	proposalsVotesRes, err := p.getProposalsVotesFromStationAPI(ctx, govPropsRes.Proposals)
	if err != nil {
		fmt.Printf("proposalsVotesRes: %v \n", err)
		return nil, nil, nil, err
	}

	return valsRes, seniorValidatorsRes, proposalsVotesRes, nil
}

func (p *allianceValidatorsProvider) getProposalsVotesFromStationAPI(ctx context.Context, proposals []*govtypes.Proposal) (stationProposals []types.StationVote, err error) {
	for _, proposal := range proposals {
		stationProposalsRes, err := p.queryStation(proposal.Id)
		if err != nil {
			fmt.Printf("stationProposalsRes: %v \n", err)
			return stationProposals, err
		}
		stationProposals = append(stationProposals, *stationProposalsRes...)
	}

	return stationProposals, err
}

func (p allianceValidatorsProvider) queryStation(propId uint64) (res *[]types.StationVote, err error) {
	// Send GET request
	resp, err := http.Get(p.stationApiUrl + "/proposals/" + fmt.Sprint(propId))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response into struct
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	// Access parsed data
	return res, nil
}

func atLeastThreeOccurrences(stationVotes []types.StationVote, val string) bool {
	count := 0
	fmt.Printf("atLeastThreeOccurrences in set %s \n", stationVotes)
	fmt.Printf("Validator %s \n", val)
	for _, v := range stationVotes {
		if v.Voter == val {
			count++
		}
	}
	return count >= 3
}
