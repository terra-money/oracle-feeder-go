package alliance_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
	types "github.com/terra-money/oracle-feeder-go/internal/types"
	pkgtypes "github.com/terra-money/oracle-feeder-go/pkg/types"
)

type alliancesQuerierProvider struct {
	feederType           types.FeederType
	transactionsProvider provider.TransactionsProvider
}

func NewAlliancesQuerierProvider(feederType types.FeederType) *alliancesQuerierProvider {
	return &alliancesQuerierProvider{
		feederType:           feederType,
		transactionsProvider: provider.NewTransactionsProvider(feederType),
	}
}

func (a alliancesQuerierProvider) QueryAndSubmitOnChain(ctx context.Context) (res []byte, err error) {
	res, err = a.requestData()
	if err != nil {
		return nil, fmt.Errorf("ERROR requesting alliances data %w", err)
	}
	txHash, err := a.transactionsProvider.SubmitAlliancesTransaction(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("ERROR submitting alliances data on chain %w", err)
	}

	fmt.Printf("Transaction Submitted successfully txHash: %s \n", txHash)
	return res, nil
}

func (a alliancesQuerierProvider) SubmitOnChain(ctx context.Context) (datxHashta string, err error) {
	var sdkMsg wasmtypes.RawContractMessage

	switch a.feederType {
	case types.AllianceHubRebalanceEmissions:
		sdkMsg, _ = json.Marshal(pkgtypes.MsgRebalanceEmissions{})
	case types.AllianceHubUpdateRewards:
		sdkMsg, _ = json.Marshal(pkgtypes.MsgUpdateRewards{})
	}

	return a.transactionsProvider.SubmitAlliancesTransaction(ctx, sdkMsg)
}

func (a alliancesQuerierProvider) requestData() (res []byte, err error) {
	var url string
	if url = os.Getenv("PRICE_SERVER_URL"); len(url) == 0 {
		url = "http://localhost:8532"
	}
	// Send GET request
	resp, err := http.Get(url + types.FromFeederTypeToPriceServerUrl(a.feederType))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Access parsed data
	return body, nil
}
