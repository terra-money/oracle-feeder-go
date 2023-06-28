package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StrideData struct {
	HostZone struct {
		ChainID           string `json:"chain_id"`
		ConnectionID      string `json:"connection_id"`
		Bech32Prefix      string `json:"bech32prefix"`
		TransferChannelID string `json:"transfer_channel_id"`
		Validators        []struct {
			Name                 string `json:"name"`
			Address              string `json:"address"`
			DelegationAmt        string `json:"delegation_amt"`
			Weight               string `json:"weight"`
			InternalExchangeRate *struct {
				InternalTokensToSharesRate string `json:"internal_tokens_to_shares_rate"`
				EpochNumber                string `json:"epoch_number"`
			} `json:"internal_exchange_rate"`
		} `json:"validators"`
		BlacklistedValidators []string `json:"blacklisted_validators"`
		WithdrawalAccount     struct {
			Address string `json:"address"`
			Target  string `json:"target"`
		} `json:"withdrawal_account"`
		FeeAccount struct {
			Address string `json:"address"`
			Target  string `json:"target"`
		} `json:"fee_account"`
		DelegationAccount struct {
			Address string `json:"address"`
			Target  string `json:"target"`
		} `json:"delegation_account"`
		RedemptionAccount struct {
			Address string `json:"address"`
			Target  string `json:"target"`
		} `json:"redemption_account"`
		IBCDenom           string  `json:"ibc_denom"`
		HostDenom          string  `json:"host_denom"`
		LastRedemptionRate sdk.Dec `json:"last_redemption_rate"`
		RedemptionRate     sdk.Dec `json:"redemption_rate"`
		UnbondingFrequency string  `json:"unbonding_frequency"`
		StakedBal          string  `json:"staked_bal"`
		Address            string  `json:"address"`
		Halted             bool    `json:"halted"`
		MinRedemptionRate  sdk.Dec `json:"min_redemption_rate"`
		MaxRedemptionRate  sdk.Dec `json:"max_redemption_rate"`
	} `json:"host_zone"`
}
