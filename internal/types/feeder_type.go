package types

import "fmt"

type FeederType string

const (
	AllianceOracleFeeder    FeederType = "alliance-oracle-feeder"
	AllianceRebalanceFeeder FeederType = "alliance-rebalance-feeder"
)

// parse from string to FeederType
func ParseFeederTypeFromString(s string) (FeederType, error) {
	switch s {
	case string(AllianceOracleFeeder):
		return AllianceOracleFeeder, nil
	case string(AllianceRebalanceFeeder):
		return AllianceRebalanceFeeder, nil
	default:
		return "", fmt.Errorf(
			`invalid feeder type: "%s", expected types are "%s" | "%s"`,
			s,
			AllianceOracleFeeder,
			AllianceRebalanceFeeder,
		)
	}
}

func FromFeederTypeToPriceServerUrl(feederType FeederType) string {
	switch feederType {
	case AllianceOracleFeeder:
		return "/alliance/protocol"
	case AllianceRebalanceFeeder:
		return "/alliance/rebalance"
	default:
		return ""
	}
}
