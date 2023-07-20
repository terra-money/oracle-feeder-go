package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StafiHubExchangeRateRes struct {
	ExchangeRate ExchangeRate `json:"exchangeRate"`
}

type ExchangeRate struct {
	Denom string  `json:"denom"`
	Value sdk.Dec `json:"value"`
}
