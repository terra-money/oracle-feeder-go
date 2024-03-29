package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/internal/provider/internal"
	"github.com/terra-money/oracle-feeder-go/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type LSDProvider struct {
	internal.BaseGrpc

	ampSTHubLuna  string
	boneSTHubLuna string

	strideApiUrl   string
	stafiHubApiUrl string

	ampSTHubWhale  string
	boneSTHubWhale string
}

func NewLSDProvider() *LSDProvider {
	return &LSDProvider{
		BaseGrpc: *internal.NewBaseGrpc(),

		ampSTHubLuna:  "terra10788fkzah89xrdm27zkj5yvhj9x3494lxawzm5qq3vvxcqz2yzaqyd3enk",
		boneSTHubLuna: "terra1l2nd99yze5fszmhl5svyh5fky9wm4nz4etlgnztfu4e8809gd52q04n3ea",

		strideApiUrl:   "https://stride-fleet.main.stridenet.co/api/Stride-Labs/stride/stakeibc/host_zone/phoenix-1",
		stafiHubApiUrl: "https://public-rest-rpc1.stafihub.io/stafihub/stafihub/ledger/exchange_rate/urswth",

		ampSTHubWhale:  "migaloo1436kxs0w2es6xlqpp9rd35e3d0cjnw4sv8j3a7483sgks29jqwgshqdky4",
		boneSTHubWhale: "migaloo1mf6ptkssddfmxvhdx0ech0k03ktp6kf9yk59renau2gvht3nq2gqdhts4u",
	}
}

func (p *LSDProvider) QueryLSTRebaseFactor(ctx context.Context, symbol string) (*sdk.Dec, error) {
	switch symbol {
	case "AMPLUNA":
		return p.queryAmpRebaseFactor(ctx, config.PHOENIX_GRPC, p.ampSTHubLuna)
	case "BACKBONELUNA":
		return p.queryBoneRebaseFactor(ctx, config.PHOENIX_GRPC, p.boneSTHubLuna)
	case "STLUNA":
		return p.queryStLunaRebaseFactor()
	case "URSWTH":
		return p.queryUrSwthRebaseFactor()
	case "AMPWHALE":
		return p.queryAmpRebaseFactor(ctx, config.MIGALOO_GRPC, p.ampSTHubWhale)
	case "BONEWHALE":
		return p.queryBoneRebaseFactor(ctx, config.MIGALOO_GRPC, p.boneSTHubWhale)
	default:
		return nil, fmt.Errorf("LSDProvider no querier implemented for symbol '%s'", symbol)
	}
}

func (p *LSDProvider) queryAmpRebaseFactor(ctx context.Context, url, address string) (*sdk.Dec, error) {
	connection, err := p.BaseGrpc.Connection(ctx, url)
	if err != nil {
		return nil, err
	}
	client := wasmtypes.NewQueryClient(connection)

	res, err := client.SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   address,
		QueryData: []byte(`{"state":{}}`),
	})
	if err != nil {
		return nil, err
	}

	var ampParsedRes types.ErisData
	err = json.Unmarshal(res.Data, &ampParsedRes)
	if err != nil {
		return nil, err
	}

	return &ampParsedRes.ExchangeRate, nil
}

func (p *LSDProvider) queryBoneRebaseFactor(ctx context.Context, url, address string) (*sdk.Dec, error) {
	connection, err := p.BaseGrpc.Connection(ctx, url)
	if err != nil {
		return nil, err
	}
	client := wasmtypes.NewQueryClient(connection)

	res, err := client.SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   address,
		QueryData: []byte(`{"state":{}}`),
	})
	if err != nil {
		return nil, err
	}

	var boneConfigParsedRes types.BoneConfigData
	err = json.Unmarshal(res.Data, &boneConfigParsedRes)
	if err != nil {
		return nil, err
	}

	return &boneConfigParsedRes.ExchangeRate, nil

}

func (p *LSDProvider) queryStLunaRebaseFactor() (*sdk.Dec, error) {
	resp, err := http.Get(p.strideApiUrl)
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

func (p *LSDProvider) queryUrSwthRebaseFactor() (*sdk.Dec, error) {
	resp, err := http.Get(p.stafiHubApiUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res types.StafiHubExchangeRateRes
	// Parse JSON response into struct
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res.ExchangeRate.Value, nil
}
