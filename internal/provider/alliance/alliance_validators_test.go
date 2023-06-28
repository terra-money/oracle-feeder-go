package alliance_provider_test

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	alliance_provider "github.com/terra-money/oracle-feeder-go/internal/provider/alliance"
	types "github.com/terra-money/oracle-feeder-go/internal/types"
)

func TestRebalanceOneVal(t *testing.T) {
	// GIVEN
	compVal := []types.ValWithAllianceTokensStake{
		{ValidatorAddr: "val1", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(10))},
		{ValidatorAddr: "val2", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(15))},
	}
	nonCompVals := []types.ValWithAllianceTokensStake{
		{ValidatorAddr: "val3", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(5))},
	}
	avgTokensPerCompVal := sdk.NewDec(15)

	// WHEN
	redelegations := alliance_provider.RebalanceVals(compVal, nonCompVals, avgTokensPerCompVal)

	// THEN
	require.Equal(t, 1, len(redelegations))
	require.Equal(t, types.Redelegation{
		SrcValidator: "val3",
		DstValidator: "val1",
		Amount:       "5",
	}, redelegations[0])
}

func TestRebalanceMultipleVal(t *testing.T) {
	// GIVEN
	compVal := []types.ValWithAllianceTokensStake{
		{ValidatorAddr: "val1", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(10))},
		{ValidatorAddr: "val2", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(2))},
		{ValidatorAddr: "val3", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(4))},
	}
	nonCompVals := []types.ValWithAllianceTokensStake{
		{ValidatorAddr: "val4", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(5))},
		{ValidatorAddr: "val5", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(5))},
		{ValidatorAddr: "val6", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(4))},
	}
	avgTokensPerCompVal := sdk.NewDec(10)

	// WHEN
	redelegations := alliance_provider.RebalanceVals(compVal, nonCompVals, avgTokensPerCompVal)
	sortByVals(redelegations)

	// THEN
	require.Equal(t, 4, len(redelegations))
	require.Equal(t, types.Redelegation{
		SrcValidator: "val4",
		DstValidator: "val2",
		Amount:       "5",
	}, redelegations[0])
	require.Equal(t, types.Redelegation{
		SrcValidator: "val5",
		DstValidator: "val2",
		Amount:       "3",
	}, redelegations[1])
	require.Equal(t, types.Redelegation{
		SrcValidator: "val5",
		DstValidator: "val3",
		Amount:       "2",
	}, redelegations[2])
	require.Equal(t, types.Redelegation{
		SrcValidator: "val6",
		DstValidator: "val3",
		Amount:       "4",
	}, redelegations[3])
}

func sortByVals(redelegations []types.Redelegation) {
	sort.Slice(redelegations, func(i, j int) bool {
		if redelegations[i].SrcValidator == redelegations[j].SrcValidator {
			return redelegations[i].DstValidator < redelegations[j].DstValidator
		}
		return redelegations[i].SrcValidator < redelegations[j].SrcValidator
	})
}
