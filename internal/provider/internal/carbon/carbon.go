package carbon

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	mintypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
)

type CarbonProvider struct {
	internal.BaseGrpc
	url   string
	denom string
}

type inflation struct {
	Result *struct {
		InflationRate *sdktypes.Dec `json:"inflationRate"`
		NumberOfWeeks *string       `json:"numberOfWeeks"`
	} `json:"result"`
}

func NewCarbonProvider() *CarbonProvider {
	return &CarbonProvider{
		url:      "https://api-insights.carbon.network/chain/inflation",
		denom:    "swth",
		BaseGrpc: *internal.NewBaseGrpc(),
	}
}

func (p *CarbonProvider) GetAnnualProvisions(ctx context.Context) (*mintypes.QueryAnnualProvisionsResponse, error) {
	grpcConn, err := p.BaseGrpc.Connection(ctx, config.CARBON_GRPC)
	if err != nil {
		return nil, err
	}
	defer grpcConn.Close()

	bankClient := banktypes.NewQueryClient(grpcConn)
	bankRes, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{
		Denom: p.denom,
	})
	if err != nil {
		return nil, err
	}

	annualInflation, err := p.getAnnualInflation()
	if err != nil {
		return nil, err
	}

	baseDenomAmount := sdktypes.NewDecWithPrec(bankRes.Amount.Amount.Int64(), 8)

	if annualInflation.Result.InflationRate == nil {
		return &mintypes.QueryAnnualProvisionsResponse{
			AnnualProvisions: sdktypes.ZeroDec(),
		}, nil
	}

	return &mintypes.QueryAnnualProvisionsResponse{
		AnnualProvisions: annualInflation.Result.InflationRate.Mul(baseDenomAmount),
	}, nil
}

func (p *CarbonProvider) getAnnualInflation() (res *inflation, err error) {
	// Send GET request
	resp, err := http.Get(p.url)
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

	return res, nil
}
