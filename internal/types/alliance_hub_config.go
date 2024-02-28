package types

type AllianceHubConfigData struct {
	Governance                string `json:"governance"`
	Controller                string `json:"controller"`
	Oracle                    string `json:"oracle"`
	LastRewardUpdateTimestamp string `json:"last_reward_update_timestamp"`
	AllianceTokenDenom        string `json:"alliance_token_denom"`
	AllianceTokenSupply       string `json:"alliance_token_supply"`
	RewardDenom               string `json:"reward_denom"`
}

type AllianceHubBalances []struct {
	Asset struct {
		Cw20   string `json:"cw20,omitempty"`
		Native string `json:"native,omitempty"`
	} `json:"asset"`
	Balance string `json:"balance,omitempty"`
}

type AllianceHubRewardDistribution []struct {
	Asset struct {
		Cw20   string `json:"cw20,omitempty"`
		Native string `json:"native,omitempty"`
	}
	Distribution string `json:"distribution,omitempty"`
}
