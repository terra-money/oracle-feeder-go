package alliance_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/terra-money/oracle-feeder-go/internal/provider"
	types "github.com/terra-money/oracle-feeder-go/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type alliancesQuerierProvider struct {
	transactionsProvider provider.TransactionsProvider
}

func NewAlliancesQuerierProvider() *alliancesQuerierProvider {
	return &alliancesQuerierProvider{
		transactionsProvider: provider.NewTransactionsProvider(),
	}
}

func (a alliancesQuerierProvider) QueryAndSubmitOnChain(ctx context.Context) (res *types.AllianceProtocolRes, err error) {
	res, err = a.RequestData()
	if err != nil {
		return nil, fmt.Errorf("ERROR requesting alliances data %w", err)
	}
	msg, err := a.transactionsProvider.ParseAlliancesTransaction(res)
	if err != nil {
		return nil, fmt.Errorf("ERROR parsing alliances data %w", err)
	}
	txHash, err := a.transactionsProvider.SubmitAlliancesTransaction(ctx, []sdk.Msg{msg})
	if err != nil {
		return nil, fmt.Errorf("ERROR submitting alliances data on chain %w", err)
	}

	fmt.Printf("Transaction Submitted successfully txHash: %s \n", txHash)
	return res, nil
}

func (alliancesQuerierProvider) RequestData() (res *types.AllianceProtocolRes, err error) {
	var url string
	if url = os.Getenv("PRICE_SERVER_URL"); len(url) == 0 {
		url = "http://localhost:8532"
	}
	// Send GET request
	resp, err := http.Get(url + "/alliance/protocol")
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

	// Access parsed data
	return res, nil
}
