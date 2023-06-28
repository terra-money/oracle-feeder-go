package config

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

type Config struct {
	Port             int                       `json:"port,omitempty"`
	MetricsPort      int                       `json:"metrics_port,omitempty"`
	Sentry           string                    `json:"sentry,omitempty"` // sentry dsn (https://sentry.io/ - error reporting service)
	Providers        map[string]ProviderConfig `json:"providers,omitempty"`
	ProviderPriority []string                  `json:"provider_prioirty,omitempty"`
}

type ProviderConfig struct {
	Symbols  []string `json:"symbols,omitempty"`
	Interval int      `json:"interval,omitempty"` // in seconds
	Timeout  int      `json:"timeout,omitempty"`
}

type AllianceConfig struct {
	GRPCUrls     []string       `json:"lcdList,omitempty"`
	LSTSData     []LSTData      `json:"lstData,omitempty"`
	LSTOnPhoenix []LSTOnPhoenix `json:"lstOnPhoenix,omitempty"`
}

type LSTData struct {
	Symbol       string
	IBCDenom     string       `json:"ibcDenom,omitempty"`
	RebaseFactor sdktypes.Dec `json:"rebaseFactor,omitempty"`
}

type LSTOnPhoenix struct {
	LSTData
	CounterpartyChainId string `json:"counterpartyChainId,omitempty"`
}
