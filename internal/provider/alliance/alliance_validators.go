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

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
)

type allianceValidatorsProvider struct {
	internal.BaseGrpc
	nodeGrpcUrl                string
	stationApiUrl              string
	allianceHubContractAddress string
	config                     *config.AllianceConfig
	providerManager            *provider.ProviderManager
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
	var allianceHubContractAddress string
	if allianceHubContractAddress = os.Getenv("ALLIANCE_HUB_CONTRACT_ADDRESS"); len(allianceHubContractAddress) == 0 {
		panic("ALLIANCE_HUB_CONTRACT_ADDRESS env variable is not set!")
	}
	return &allianceValidatorsProvider{
		BaseGrpc:                   *internal.NewBaseGrpc(),
		config:                     config,
		nodeGrpcUrl:                nodeGrpcUrl,
		stationApiUrl:              stationApiUrl,
		providerManager:            providerManager,
		allianceHubContractAddress: allianceHubContractAddress,
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

	smartContractRes, err := p.querySmartContractConfig(ctx)
	if err != nil {
		return nil, err
	}

	stakingValidators, seniorValidators, proposalsVotes, allianceVals, err := p.queryValidatorsData(ctx)
	if err != nil {
		return nil, err
	}

	var compliantVals []stakingtypes.Validator
	// Apply the previous rules to filter the list
	// of all blockchain validators to keep the ones
	// that are compliant
	for _, val := range stakingValidators {
		// (1) skip if status is not bonded (again in case the api have a bug with the query)
		if val.GetStatus() != stakingtypes.Bonded {
			continue
		}

		// (2) skip if jailed
		if val.IsJailed() {
			continue
		}

		// (3) skip if commission is grater than 19%
		if val.Commission.CommissionRates.Rate.GT(sdktypes.MustNewDecFromStr("0.19")) {
			continue
		}

		// (4) skip if have not voted in the last 3 proposals
		if !atLeastThreeOccurrences(proposalsVotes, val.OperatorAddress) {
			continue
		}

		// (5) skip if it have not been in the active validator set 100 000 blocks before the current one
		for _, seniorValidator := range seniorValidators {
			if val.OperatorAddress != seniorValidator.Address {
				continue
			}
		}

		compliantVals = append(compliantVals, val)
	}

	valsWithAllianceTokens, totalAllianceStakedTokens := p.filterAllianceValsWithStake(allianceVals, smartContractRes)
	compliantValsWithAllianceTokens,
		nonCompliantValsWithAllianceTokens,
		totalAllianceStakedTokensOnNonCompliantVals := p.filterAllianceValsByCompliance(compliantVals, valsWithAllianceTokens)
	avgTokensPerCompliantVal := totalAllianceStakedTokens.Quo(sdktypes.NewDec(int64(len(compliantValsWithAllianceTokens))))

	redelegations := p.rebalanceVals(
		compliantValsWithAllianceTokens,
		nonCompliantValsWithAllianceTokens,
		avgTokensPerCompliantVal,
		totalAllianceStakedTokensOnNonCompliantVals,
	)

	return types.NewAllianceRedelegateReq(redelegations), nil
}

// In charge of rebalancing the stake from non-compliant validators to compliant ones.
// - Compliant validators shouldn't have more than the average amout of stake (avgTokensPerCompVal).
// - Non-compliant validators should end with 0 stake at the end of function execution.
// - If any compliant validator has more than the average amout of stake, re-balance to other compliant validators.
func (*allianceValidatorsProvider) rebalanceVals(
	compVal []types.ValWithAllianceTokensStake,
	nonCompVals []types.ValWithAllianceTokensStake,
	avgTokensPerCompVal sdktypes.Dec,
	tokensStakedOnNonCompVals sdktypes.Dec,
) []types.Redelegation {
	redelegations := []types.Redelegation{}

	return redelegations
}

// Method to split the list of alliance validators in two subsets:
//   - compliantValsWithAllianceTokens: comply with the rules described below the method GetAllianceRedelegateReq,
//   - nonCompliantValsWithAllianceTokens: the ones that does not complie with the rules,
//
// It also counts how much stake has been delegated to the non
// compliant validators and returns the amout of stake that should be rebalanced
func (*allianceValidatorsProvider) filterAllianceValsByCompliance(compliantVals []stakingtypes.Validator, valsWithAllianceTokens []types.ValWithAllianceTokensStake) ([]types.ValWithAllianceTokensStake, []types.ValWithAllianceTokensStake, sdktypes.Dec) {
	compliantValsWithAllianceTokens := []types.ValWithAllianceTokensStake{}
	nonCompliantValsWithAllianceTokens := []types.ValWithAllianceTokensStake{}
	totalAllianceStakedTokensOnNonCompliantVals := sdktypes.ZeroDec()

	for _, valWithAllianceTokensStake := range valsWithAllianceTokens {
		found := false

		for _, val := range compliantVals {

			if val.OperatorAddress == valWithAllianceTokensStake.ValidatorAddr {
				compliantValsWithAllianceTokens = append(
					compliantValsWithAllianceTokens,
					valWithAllianceTokensStake,
				)
				found = true
				continue
			}
		}
		if !found {
			nonCompliantValsWithAllianceTokens = append(
				nonCompliantValsWithAllianceTokens,
				valWithAllianceTokensStake,
			)
			totalAllianceStakedTokensOnNonCompliantVals = totalAllianceStakedTokensOnNonCompliantVals.Add(valWithAllianceTokensStake.TotalStaked.Amount)
		}
	}

	return compliantValsWithAllianceTokens, nonCompliantValsWithAllianceTokens, totalAllianceStakedTokensOnNonCompliantVals
}

// Filter the alliance validators to keep only the ones that have staked ualliance tokens
func (*allianceValidatorsProvider) filterAllianceValsWithStake(allianceVals []alliancetypes.QueryAllianceValidatorResponse, smartContractRes *types.AllianceHubConfigData) ([]types.ValWithAllianceTokensStake, sdktypes.Dec) {
	valsWithAllianceTokens := []types.ValWithAllianceTokensStake{}
	uallianceStakedTokens := sdktypes.ZeroDec()

	for _, val := range allianceVals {
		for _, stake := range val.TotalStaked {
			// As soon as we find the first entry with the alliance token denom
			// we can append the validator to the list and break the loop
			// because it represents all the ualliance tokens staked to that validator
			if stake.Denom == smartContractRes.AllianceTokenDenom {
				valsWithAllianceTokens = append(
					valsWithAllianceTokens,
					types.NewValWithAllianceTokensStake(val.ValidatorAddr, stake),
				)

				uallianceStakedTokens = uallianceStakedTokens.Add(stake.Amount)
				continue
			}
		}
	}
	return valsWithAllianceTokens, uallianceStakedTokens
}

func (p *allianceValidatorsProvider) querySmartContractConfig(ctx context.Context) (*types.AllianceHubConfigData, error) {
	grpcConn, err := p.BaseGrpc.Connection(p.nodeGrpcUrl, nil)
	if err != nil {
		fmt.Printf("grpcConn: %v \n", err)
		return nil, err
	}
	defer grpcConn.Close()
	client := wasmtypes.NewQueryClient(grpcConn)

	res, err := client.SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   p.allianceHubContractAddress,
		QueryData: []byte(`{ "config" : {}}`),
	})
	if err != nil {
		return nil, err
	}

	var configRes types.AllianceHubConfigData
	err = json.Unmarshal(res.Data, &configRes)
	if err != nil {
		return nil, err
	}

	return &configRes, nil
}

func (p *allianceValidatorsProvider) queryValidatorsData(ctx context.Context) (
	[]stakingtypes.Validator,
	[]*tmservice.Validator,
	[]types.StationVote,
	[]alliancetypes.QueryAllianceValidatorResponse,
	error,
) {
	grpcConn, err := p.BaseGrpc.Connection(p.nodeGrpcUrl, nil)
	if err != nil {
		fmt.Printf("grpcConn: %v \n", err)
		return nil, nil, nil, nil, err
	}
	defer grpcConn.Close()

	nodeClient := tmservice.NewServiceClient(grpcConn)
	govClient := govtypes.NewQueryClient(grpcConn)
	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	allianceClient := alliancetypes.NewQueryClient(grpcConn)

	valsRes, err := stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded, // (1) query only status bonded
		Pagination: &query.PageRequest{
			Limit: 150,
		},
	})
	if err != nil {
		fmt.Printf("valsRes: %v \n", err)
		return nil, nil, nil, nil, err
	}

	govPropsRes, err := govClient.Proposals(ctx, &govtypes.QueryProposalsRequest{
		Pagination: &query.PageRequest{
			Limit:   3,
			Reverse: true,
		},
	})
	if err != nil {
		fmt.Printf("govPropsRes: %v \n", err)
		return nil, nil, nil, nil, err
	}

	latestHeightRes, err := nodeClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		fmt.Printf("latestHeightRes: %v \n", err)
		return nil, nil, nil, nil, err
	}

	seniorValidatorsRes, err := nodeClient.GetValidatorSetByHeight(ctx, &tmservice.GetValidatorSetByHeightRequest{
		Height: latestHeightRes.SdkBlock.Header.Height - 100_000,
	})
	if err != nil {
		fmt.Printf("seniorValidatorsRes: %v \n", err)
		return nil, nil, nil, nil, err
	}

	proposalsVotesRes, err := p.getProposalsVotesFromStationAPI(ctx, govPropsRes.Proposals)
	if err != nil {
		fmt.Printf("proposalsVotesRes: %v \n", err)
		return nil, nil, nil, nil, err
	}

	allianceVals, err := allianceClient.AllAllianceValidators(ctx, &alliancetypes.QueryAllAllianceValidatorsRequest{
		Pagination: &query.PageRequest{
			Limit: 150,
		},
	})
	if err != nil {
		fmt.Printf("allianceVals: %v \n", err)
		return nil, nil, nil, nil, err
	}

	return valsRes.Validators,
		seniorValidatorsRes.Validators,
		proposalsVotesRes,
		allianceVals.Validators,
		nil
}

func (p *allianceValidatorsProvider) getProposalsVotesFromStationAPI(ctx context.Context, proposals []*govtypes.Proposal) (stationProposals []types.StationVote, err error) {
	for _, proposal := range proposals {
		stationProposalsRes, err := p.queryStation(proposal.Id)
		if err != nil {
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
	for _, v := range stationVotes {
		if v.Voter == val {
			count++
		}
	}
	return count >= 3
}
