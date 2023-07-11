package types

import "fmt"

type FeederType string

const (
	AllianceHubUpdateRewards      FeederType = "alliance-hub-update-rewards"
	AllianceHubRebalanceEmissions FeederType = "Alliance-hub-rebalance-emissions"
	AllianceOracleFeeder          FeederType = "alliance-oracle-feeder"
	AllianceRebalanceFeeder       FeederType = "alliance-rebalance-feeder"
)

// parse from string to FeederType
func ParseFeederTypeFromString(s string) (FeederType, error) {
	switch s {
	case string(AllianceOracleFeeder):
		return AllianceOracleFeeder, nil
	case string(AllianceRebalanceFeeder):
		return AllianceRebalanceFeeder, nil
	case string(AllianceHubUpdateRewards):
		return AllianceHubUpdateRewards, nil
	case string(AllianceHubRebalanceEmissions):
		return AllianceHubRebalanceEmissions, nil
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
