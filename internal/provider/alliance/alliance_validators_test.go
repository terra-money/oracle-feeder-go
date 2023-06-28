package alliance_provider_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
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

func TestRebalanceMultipleVal2(t *testing.T) {
	// GIVEN
	compVal := []types.ValWithAllianceTokensStake{
		{ValidatorAddr: "val1", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(40))},
		{ValidatorAddr: "val2", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(2))},
		{ValidatorAddr: "val3", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(4))},
	}
	nonCompVals := []types.ValWithAllianceTokensStake{
		{ValidatorAddr: "val4", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(5))},
		{ValidatorAddr: "val5", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(5))},
		{ValidatorAddr: "val6", TotalStaked: sdk.NewDecCoinFromDec("token", sdk.NewDec(4))},
	}
	avgTokensPerCompVal := sdk.NewDec(20)

	// WHEN
	redelegations := alliance_provider.RebalanceVals(compVal, nonCompVals, avgTokensPerCompVal)

	// THEN
	require.Equal(t, 5, len(redelegations))
	require.Equal(t, types.Redelegation{
		SrcValidator: "val4",
		DstValidator: "val2",
		Amount:       "5",
	}, redelegations[0])
	require.Equal(t, types.Redelegation{
		SrcValidator: "val5",
		DstValidator: "val2",
		Amount:       "5",
	}, redelegations[1])
	require.Equal(t, types.Redelegation{
		SrcValidator: "val6",
		DstValidator: "val2",
		Amount:       "4",
	}, redelegations[2])
	require.Equal(t, types.Redelegation{
		SrcValidator: "val1",
		DstValidator: "val2",
		Amount:       "4",
	}, redelegations[3])
	require.Equal(t, types.Redelegation{
		SrcValidator: "val1",
		DstValidator: "val3",
		Amount:       "16",
	}, redelegations[4])
}
func TestFilterAllianceValsWithStake(t *testing.T) {
	// GIVEN
	allianceVals := []alliancetypes.QueryAllianceValidatorResponse{
		{
			ValidatorAddr: "val1",
			TotalStaked: []sdk.DecCoin{
				{Denom: "token1", Amount: sdk.NewDec(100)},
			},
		},
		{
			ValidatorAddr: "val2",
			TotalStaked: []sdk.DecCoin{
				{Denom: "token1", Amount: sdk.NewDec(300)},
				{Denom: "token2", Amount: sdk.NewDec(300)},
			},
		},
		{
			ValidatorAddr: "val3",
			TotalStaked: []sdk.DecCoin{
				{Denom: "token2", Amount: sdk.NewDec(300)},
			},
		},
	}
	allianceTokenDenom := "token1"

	// WHEN
	valsWithAllianceTokens, uallianceStakedTokens := alliance_provider.FilterAllianceValsWithStake(allianceVals, allianceTokenDenom)

	// THEN
	require.Equal(t, sdk.NewDec(400), uallianceStakedTokens)
	require.Equal(t, []types.ValWithAllianceTokensStake{
		{
			ValidatorAddr: "val1",
			TotalStaked:   sdk.DecCoin{Denom: "token1", Amount: sdk.NewDec(100)},
		},
		{
			ValidatorAddr: "val2",
			TotalStaked:   sdk.DecCoin{Denom: "token1", Amount: sdk.NewDec(300)},
		},
	}, valsWithAllianceTokens)
}

func TestParseAllianceValsByCompliance(t *testing.T) {
	// GIVEN
	compliantVals := []stakingtypes.Validator{
		{
			OperatorAddress: "val1",
		},
		{
			OperatorAddress: "val2",
		},
		{
			OperatorAddress: "val3",
		},
	}
	valsWithAllianceTokens := []types.ValWithAllianceTokensStake{
		{
			ValidatorAddr: "val1",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(100)),
		},
		{
			ValidatorAddr: "val2",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(200)),
		},
		{
			ValidatorAddr: "val4",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(200)),
		},
	}
	allianceTokenDenom := "token"

	// WHEN
	compliant, nonCompliant := alliance_provider.ParseAllianceValsByCompliance(compliantVals, valsWithAllianceTokens, allianceTokenDenom)

	// THEN
	require.Equal(t, []types.ValWithAllianceTokensStake{
		{
			ValidatorAddr: "val1",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(100)),
		},
		{
			ValidatorAddr: "val2",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(200)),
		},
		{
			ValidatorAddr: "val3",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(0)),
		},
	}, compliant)
	require.Equal(t, []types.ValWithAllianceTokensStake{

		{
			ValidatorAddr: "val4",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(200)),
		},
	}, nonCompliant)
}

func TestParseAllianceValsByComplianceWithNoStake(t *testing.T) {
	// GIVEN
	compliantVals := []stakingtypes.Validator{
		{
			OperatorAddress: "val1",
		},
		{
			OperatorAddress: "val2",
		},
		{
			OperatorAddress: "val3",
		},
	}
	valsWithAllianceTokens := []types.ValWithAllianceTokensStake{}
	allianceTokenDenom := "token"

	// WHEN
	compliant, nonCompliant := alliance_provider.ParseAllianceValsByCompliance(compliantVals, valsWithAllianceTokens, allianceTokenDenom)

	// THEN
	require.Equal(t, []types.ValWithAllianceTokensStake{
		{
			ValidatorAddr: "val1",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(0)),
		},
		{
			ValidatorAddr: "val2",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(0)),
		},
		{
			ValidatorAddr: "val3",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(0)),
		},
	}, compliant)
	require.Equal(t, []types.ValWithAllianceTokensStake{}, nonCompliant)

}

func TestParseAllianceValsWithNoneCompliant(t *testing.T) {
	// GIVEN
	compliantVals := []stakingtypes.Validator{}
	valsWithAllianceTokens := []types.ValWithAllianceTokensStake{
		{
			ValidatorAddr: "val1",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(100)),
		},
		{
			ValidatorAddr: "val2",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(200)),
		},
		{
			ValidatorAddr: "val3",
			TotalStaked:   sdk.NewDecCoinFromDec("token", sdk.NewDec(0)),
		}}
	allianceTokenDenom := "token"

	// WHEN
	compliant, nonCompliant := alliance_provider.ParseAllianceValsByCompliance(compliantVals, valsWithAllianceTokens, allianceTokenDenom)

	// THEN
	require.Equal(t, 0, len(compliant))
	require.Equal(t, 3, len(nonCompliant))

}
