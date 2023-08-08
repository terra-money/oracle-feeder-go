package types

import "fmt"

type FeederType string

const (
	AllianceInitialDelegation  FeederType = "alliance-initial-delegation"
	AllianceUpdateRewards      FeederType = "alliance-update-rewards"
	AllianceRebalanceEmissions FeederType = "alliance-rebalance-emissions"
	AllianceOracleFeeder       FeederType = "alliance-oracle-feeder"
	AllianceRebalanceFeeder    FeederType = "alliance-rebalance-feeder"
)

// parse from string to FeederType
func ParseFeederTypeFromString(s string) (FeederType, error) {
	switch s {
	case string(AllianceInitialDelegation):
		return AllianceInitialDelegation, nil
	case string(AllianceUpdateRewards):
		return AllianceUpdateRewards, nil
	case string(AllianceRebalanceEmissions):
		return AllianceRebalanceEmissions, nil
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
	case AllianceInitialDelegation:
		return "/alliance/rebalance"
	default:
		return ""
	}
}
