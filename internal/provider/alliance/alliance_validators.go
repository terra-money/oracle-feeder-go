package alliance_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
	types "github.com/terra-money/oracle-feeder-go/internal/types"
	pkgtypes "github.com/terra-money/oracle-feeder-go/pkg/types"

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
	config                             *config.AllianceConfig
	providerManager                    *provider.ProviderManager
	nodeGrpcUrl                        string
	stationApiUrl                      string
	allianceHubContractAddress         string
	blocksToBeSeniorValidator          int64
	voteOnProposalsToBeSeniorValidator int64
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

	var blocksToBeSeniorValidator int64 = 100_000
	if blocks := os.Getenv("BLOCKS_TO_BE_SENIOR_VALIDATOR"); len(blocks) != 0 {
		blocks, err := strconv.ParseInt(blocks, 10, 64)
		if err != nil {
			panic("BLOCKS_TO_BE_SENIOR_VALIDATOR env variable is not a valid integer!")
		}

		blocksToBeSeniorValidator = blocks
	}

	var voteOnProposalsToBeSeniorValidator int64 = 3
	if proposals := os.Getenv("VOTE_ON_PROPOSALS_TO_BE_SENIOR_VALIDATOR"); len(proposals) != 0 {
		voteOnProposals, err := strconv.ParseInt(proposals, 10, 64)
		if err != nil {
			panic("VOTE_ON_PROPOSALS_TO_BE_SENIOR_VALIDATOR env variable is not a valid integer!")
		}
		if voteOnProposals > 3 {
			panic("VOTE_ON_PROPOSALS_TO_BE_SENIOR_VALIDATOR env variable is greater than 3!")
		}
		voteOnProposalsToBeSeniorValidator = voteOnProposals
	}

	return &allianceValidatorsProvider{
		BaseGrpc:                           *internal.NewBaseGrpc(),
		config:                             config,
		nodeGrpcUrl:                        nodeGrpcUrl,
		stationApiUrl:                      stationApiUrl,
		providerManager:                    providerManager,
		allianceHubContractAddress:         allianceHubContractAddress,
		blocksToBeSeniorValidator:          blocksToBeSeniorValidator,
		voteOnProposalsToBeSeniorValidator: voteOnProposalsToBeSeniorValidator,
	}
}

// Query Terra GRPC, station API and return the list of alliance
// redelegations that can be submitted to Alliance Hub applying
// the following rules:
// (1) - be part of the active validator set,
// (2) - to do not be jailed,
// (3) - commission rate to be lower than 10%,
// (4) - Participate in the latest 3 gov proposals,
// (5) - have been in the active validator set 100 000 blocks before the current one, (1 week approx)
func (p *allianceValidatorsProvider) GetAllianceInitialDelegations(ctx context.Context) (*pkgtypes.MsgAllianceDelegations, error) {
	smartContractRes, err := p.querySmartContractConfig(ctx)
	if err != nil {
		return nil, err
	}

	stakingValidators, seniorValidators, proposalsVotes, _, err := p.queryValidatorsData(ctx)
	if err != nil {
		return nil, err
	}
	compliantVals := p.getCompliantValidators(stakingValidators, proposalsVotes, seniorValidators)
	allianceTokenSupply, err := strconv.ParseInt(smartContractRes.AllianceTokenSupply, 10, 64)
	if err != nil {
		panic(err)
	}
	tokensPerVal := allianceTokenSupply / int64(len(compliantVals))

	var delegations []types.Delegation
	for i := 0; i < len(compliantVals); i++ {
		delegations = append(
			delegations,
			types.NewDelegation(
				compliantVals[i].OperatorAddress,
				fmt.Sprint(tokensPerVal),
			),
		)
	}

	res := pkgtypes.NewMsgAllianceDelegations(delegations)

	return &res, nil
}

// Query Terra GRPC, station API and return the list of alliance
// redelegations that can be submitted to Alliance Hub applying
// the following rules:
// (1) - be part of the active validator set,
// (2) - to do not be jailed,
// (3) - commission rate to be lower than 10%,
// (4) - Participate in the latest 3 gov proposals,
// (5) - have been in the active validator set 100 000 blocks before the current one, (1 week approx)
func (p *allianceValidatorsProvider) GetAllianceRedelegateReq(ctx context.Context) (*pkgtypes.MsgAllianceRedelegate, error) {
	smartContractRes, err := p.querySmartContractConfig(ctx)
	if err != nil {
		return nil, err
	}

	stakingValidators, seniorValidators, proposalsVotes, allianceVals, err := p.queryValidatorsData(ctx)
	if err != nil {
		return nil, err
	}

	// Apply the previous rules to filter the list of validators
	compliantVals := p.getCompliantValidators(stakingValidators, proposalsVotes, seniorValidators)

	valsWithAllianceTokens, totalAllianceStakedTokens := FilterAllianceValsWithStake(allianceVals, smartContractRes.AllianceTokenDenom)
	compliantValsWithAllianceTokens,
		nonCompliantValsWithAllianceTokens := ParseAllianceValsByCompliance(compliantVals, valsWithAllianceTokens, smartContractRes.AllianceTokenDenom)
	avgTokensPerCompliantVal := totalAllianceStakedTokens.Quo(sdktypes.NewDec(int64(len(compliantVals))))

	redelegations := RebalanceVals(
		compliantValsWithAllianceTokens,
		nonCompliantValsWithAllianceTokens,
		avgTokensPerCompliantVal,
	)

	res := pkgtypes.NewMsgAllianceRedelegate(redelegations)

	return &res, nil
}

// Apply the rules to filter the list of validators
// (1) - be part of the active validator set,
// (2) - to do not be jailed,
// (3) - commission rate to be lower than 10%,
// (4) - Participate in the latest 3 gov proposals,
// (5) - have been in the active validator set 100 000 blocks before the current one, (1 week approx)
func (p *allianceValidatorsProvider) getCompliantValidators(stakingValidators []stakingtypes.Validator, proposalsVotes []types.StationVote, seniorValidators []*tmservice.Validator) []stakingtypes.Validator {
	var compliantVals []stakingtypes.Validator

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

		// (3) skip if commission is grater than 10%
		if val.Commission.CommissionRates.Rate.GT(sdktypes.MustNewDecFromStr("0.10")) {
			continue
		}

		// (4) skip if have not voted in the last 3 proposals
		if !p.votedForLatestProposals(proposalsVotes, val.OperatorAddress) {
			continue
		}

		for _, seniorValidator := range seniorValidators {
			// (5) skip if it have not been in the active validator set 100 000 blocks before the current one
			if val.OperatorAddress != seniorValidator.Address {
				continue
			}
		}

		compliantVals = append(compliantVals, val)
	}
	return compliantVals
}

// In charge of rebalancing the stake from non-compliant validators to compliant ones.
// - Non-compliant validators should end with 0 stake at the end of function execution.
// - Compliant validators shouldn't have more than the average amout of stake (avgTokensPerComplVal).
// - If any compliant validator has more than the average amout of stake, re-balance to other compliant validators.
func RebalanceVals(
	compVal []types.ValWithAllianceTokensStake,
	nonCompVals []types.ValWithAllianceTokensStake,
	avgTokensPerComplVal sdktypes.Dec,
) []types.Redelegation {
	redelegations := []types.Redelegation{}

	// Redelegate the non-compliant validators stake
	// until they have 0 stake.
	for i := 0; i < len(nonCompVals); i++ {

		for j := 0; j < len(compVal); j++ {
			nonCompValStake := nonCompVals[i].TotalStaked.Amount
			if nonCompValStake.IsZero() {
				break
			}
			compValStake := compVal[j].TotalStaked.Amount
			if compValStake.LT(avgTokensPerComplVal) {
				// ... calculate the delta to the average
				deltaStakeToRebalance := avgTokensPerComplVal.Sub(compValStake)

				// If the delta is greater than the stake of the non-compliant validator
				// use all the stake of the non-compliant validator
				if deltaStakeToRebalance.GT(nonCompValStake) {
					deltaStakeToRebalance = nonCompValStake
				}

				// Append the redelegation to the list
				redelegations = append(
					redelegations,
					types.NewRedelegation(
						nonCompVals[i].ValidatorAddr,
						compVal[j].ValidatorAddr,
						// Since the operations are done with Decimals, we need to remove the decimal part https://github.com/terra-money/alliance/issues/227
						strings.Split(deltaStakeToRebalance.String(), ".")[0],
					),
				)

				// Update the stake of the compliant validator
				compVal[j].TotalStaked.Amount = compValStake.Add(deltaStakeToRebalance)
				// Update the stake of the non-compliant validator
				nonCompVals[i].TotalStaked.Amount = nonCompValStake.Sub(deltaStakeToRebalance)
			}
		}
	}

	// Redelegate the compliant validators stake
	// if any has more than the average stake.
	for i := 0; i < len(compVal); i++ {
		for j := 0; j < len(compVal); j++ {
			IcompValStake := compVal[i].TotalStaked.Amount
			// break if the src validator has less than or equal to the average stake
			if IcompValStake.LTE(avgTokensPerComplVal) {
				break
			}
			JcompValStake := compVal[j].TotalStaked.Amount
			if JcompValStake.LT(avgTokensPerComplVal) {
				// ... and calculate the delta to the average
				diffNeededByDstVal := avgTokensPerComplVal.Sub(JcompValStake)
				diffAvailableBySrcVal := IcompValStake.Sub(avgTokensPerComplVal)

				// Take the minimum between the two
				var deltaStakeToRebalance sdktypes.Dec
				if diffNeededByDstVal.GT(diffAvailableBySrcVal) {
					deltaStakeToRebalance = diffAvailableBySrcVal
				} else {
					deltaStakeToRebalance = diffNeededByDstVal
				}

				// Append the redelegation to the list
				redelegations = append(
					redelegations,
					types.NewRedelegation(
						compVal[i].ValidatorAddr,
						compVal[j].ValidatorAddr,
						// Since the operations are done with Decimals, we need to remove the decimal part https://github.com/terra-money/alliance/issues/227
						strings.Split(deltaStakeToRebalance.String(), ".")[0],
					),
				)

				// Update the stake of the validator with more stake than the average
				compVal[j].TotalStaked.Amount = JcompValStake.Add(deltaStakeToRebalance)
				// Update the stake of the validator with less stake than the average
				compVal[i].TotalStaked.Amount = IcompValStake.Sub(deltaStakeToRebalance)
			}
		}
	}

	return redelegations
}

// Method to split the list of alliance validators in two subsets:
//
//   - compliantValsWithAllianceTokens: validatos that comply with the rules
//     described below the method GetAllianceRedelegateReq with the stake to zero if has no stake.
//
//   - nonCompliantValsWithAllianceTokens: the ones that does not complie with the rules
//     in this subset should never exist validators with zero stake,
func ParseAllianceValsByCompliance(
	compliantVals []stakingtypes.Validator,
	valsWithAllianceTokens []types.ValWithAllianceTokensStake,
	allianceTokenDenom string,
) ([]types.ValWithAllianceTokensStake, []types.ValWithAllianceTokensStake) {
	compliantValsWithAllianceTokens := []types.ValWithAllianceTokensStake{}
	nonCompliantValsWithAllianceTokens := []types.ValWithAllianceTokensStake{}

	// Parse the **compliantVals** to the type **ValWithAllianceTokensStake**
	// if they have no stake initialize the stake to 0.
	for _, val := range compliantVals {
		compliantValsWithAllianceTokens = append(
			compliantValsWithAllianceTokens,
			types.ValWithAllianceTokensStake{
				ValidatorAddr: val.OperatorAddress,
				TotalStaked:   sdktypes.NewDecCoinFromDec(allianceTokenDenom, sdktypes.ZeroDec()),
			},
		)
	}

	// Update the stake of the compliant validators since it was initialized to 0
	// and populate the list of the non-compliant validators.
	for _, valWithAllianceTokensStake := range valsWithAllianceTokens {
		found := false
		for i := 0; i < len(compliantValsWithAllianceTokens); i++ {
			if compliantValsWithAllianceTokens[i].ValidatorAddr == valWithAllianceTokensStake.ValidatorAddr {
				compliantValsWithAllianceTokens[i].TotalStaked = valWithAllianceTokensStake.TotalStaked
				found = true
				continue
			}
		}

		if !found {
			nonCompliantValsWithAllianceTokens = append(
				nonCompliantValsWithAllianceTokens,
				valWithAllianceTokensStake,
			)
		}
	}

	return compliantValsWithAllianceTokens, nonCompliantValsWithAllianceTokens
}

// Filter the alliance validators to keep only the ones that have staked ualliance tokens
func FilterAllianceValsWithStake(allianceVals []alliancetypes.QueryAllianceValidatorResponse, allianceTokenDenom string) ([]types.ValWithAllianceTokensStake, sdktypes.Dec) {
	valsWithAllianceTokens := []types.ValWithAllianceTokensStake{}
	uallianceStakedTokens := sdktypes.ZeroDec()

	for _, val := range allianceVals {
		for _, stake := range val.TotalStaked {
			// As soon as we find the first entry with the alliance token denom
			// we can append the validator to the list and break the loop
			// because it represents all the ualliance tokens staked to that validator
			if stake.Denom == allianceTokenDenom {
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
	grpcConn, err := p.BaseGrpc.Connection(ctx, p.nodeGrpcUrl)
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
	grpcConn, err := p.BaseGrpc.Connection(ctx, p.nodeGrpcUrl)
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
		Height: latestHeightRes.SdkBlock.Header.Height - p.blocksToBeSeniorValidator,
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
	url := p.stationApiUrl + "/proposals/" + fmt.Sprint(propId)
	// Send GET request
	resp, err := http.Get(url)
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

func (p allianceValidatorsProvider) votedForLatestProposals(stationVotes []types.StationVote, val string) bool {
	count := 0
	for _, v := range stationVotes {
		if v.Voter == val {
			count++
		}
	}
	return count >= int(p.voteOnProposalsToBeSeniorValidator)
}
