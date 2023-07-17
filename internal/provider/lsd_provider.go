package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
	"github.com/terra-money/oracle-feeder-go/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type LSDProvider struct {
	internal.BaseGrpc
	phoenixNodeUrl                string
	striddeApiUrl                 string
	erisStakingHubContractAddress string
	boneLunaHubContractAddress    string
}

func NewLSDProvider() *LSDProvider {
	return &LSDProvider{
		BaseGrpc:                      *internal.NewBaseGrpc(),
		phoenixNodeUrl:                "terra-grpc.polkachu.com:11790",
		striddeApiUrl:                 "https://stride-fleet.main.stridenet.co/api/Stride-Labs/stride/stakeibc/host_zone/phoenix-1",
		erisStakingHubContractAddress: "terra10788fkzah89xrdm27zkj5yvhj9x3494lxawzm5qq3vvxcqz2yzaqyd3enk",
		boneLunaHubContractAddress:    "terra1l2nd99yze5fszmhl5svyh5fky9wm4nz4etlgnztfu4e8809gd52q04n3ea",
	}
}

func (p *LSDProvider) QueryLSTRebaseFactor(symbol string) (*sdk.Dec, error) {
	switch symbol {
	case "AMPLUNA":
		return p.queryAmpLunaRebaseFactor()
	case "BACKBONELUNA":
		return p.queryBoneLunaRebaseFactor()
	case "STLUNA":
		return p.queryStLunaRebaseFactor()
	default:
		return nil, fmt.Errorf("LSDProvider no querier implemented for symbol '%s'", symbol)
	}
}

func (p *LSDProvider) queryAmpLunaRebaseFactor() (*sdk.Dec, error) {
	ctx := context.Background()
	connection, err := p.BaseGrpc.Connection(ctx, p.phoenixNodeUrl)
	if err != nil {
		return nil, err
	}
	client := wasmtypes.NewQueryClient(connection)

	res, err := client.SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   p.erisStakingHubContractAddress,
		QueryData: []byte(`{ "state" : {}}`),
	})
	if err != nil {
		return nil, err
	}

	var erisParsedRes types.ErisData
	err = json.Unmarshal(res.Data, &erisParsedRes)
	if err != nil {
		return nil, err
	}

	return &erisParsedRes.ExchangeRate, nil
}

func (p *LSDProvider) queryBoneLunaRebaseFactor() (*sdk.Dec, error) {
	ctx := context.Background()
	connection, err := p.BaseGrpc.Connection(ctx, p.phoenixNodeUrl)
	if err != nil {
		return nil, err
	}
	client := wasmtypes.NewQueryClient(connection)

	res, err := client.SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   p.boneLunaHubContractAddress,
		QueryData: []byte(`{ "state" : {}}`),
	})
	if err != nil {
		return nil, err
	}

	var boneLunaParsedRes types.BoneLunaData
	err = json.Unmarshal(res.Data, &boneLunaParsedRes)
	if err != nil {
		return nil, err
	}

	return &boneLunaParsedRes.ExchangeRate, nil

}

func (p *LSDProvider) queryStLunaRebaseFactor() (*sdk.Dec, error) {
	resp, err := http.Get(p.striddeApiUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res types.StrideData
	// Parse JSON response into struct
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res.HostZone.RedemptionRate, nil
}
