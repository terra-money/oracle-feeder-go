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
	telegramProvider     *provider.TelegramProvider
}

func NewAlliancesQuerierProvider(feederType types.FeederType) *alliancesQuerierProvider {
	return &alliancesQuerierProvider{
		telegramProvider:     provider.NewTelegramProvider(),
		feederType:           feederType,
		transactionsProvider: provider.NewTransactionsProvider(feederType),
	}
}

func (a alliancesQuerierProvider) SubmitTx(ctx context.Context) (hash string, err error) {
	if a.feederType == types.AllianceOracleFeeder ||
		a.feederType == types.AllianceRebalanceFeeder ||
		a.feederType == types.AllianceInitialDelegation {
		hash, err = a.QueryAndSubmitOnChain(ctx)
	} else {
		hash, err = a.SubmitOnChain(ctx)
	}

	if err != nil {
		_ = a.telegramProvider.SendError(string(a.feederType), err)
		return "", err
	}

	_ = a.sendSuccessTelegramMessage(hash)

	return hash, err
}

func (a alliancesQuerierProvider) QueryAndSubmitOnChain(ctx context.Context) (string, error) {
	res, err := a.requestData()
	if err != nil {
		return "", fmt.Errorf("ERROR querying alliances data %w", err)
	}
  
	return a.transactionsProvider.SubmitAlliancesTransaction(ctx, res)
}

func (a alliancesQuerierProvider) SubmitOnChain(ctx context.Context) (string, error) {
	var sdkMsg wasmtypes.RawContractMessage

	switch a.feederType {
	case types.AllianceRebalanceEmissions:
		sdkMsg, _ = json.Marshal(pkgtypes.MsgRebalanceEmissions{})
	case types.AllianceUpdateRewards:
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

func (a alliancesQuerierProvider) sendSuccessTelegramMessage(hash string) error {
	url := fmt.Sprintf("<a href='https://finder.terra.money/%s/tx/%s'>(transaction)</a>", a.transactionsProvider.ChainId, hash)
	var msg string

	switch a.feederType {
	case types.AllianceRebalanceEmissions:
		msg = fmt.Sprintf("<b>[Alliance Hub]</b> Staking rewards rebalanced successfully %s", url)
	case types.AllianceUpdateRewards:
		msg = fmt.Sprintf("<b>[Alliance Hub]</b> Staking rewards updated successfully %s", url)
	case types.AllianceInitialDelegation:
		msg = fmt.Sprintf("<b>[Alliance Hub]</b> Initial delegation of `ualliance` tokens successfully %s", url)
	case types.AllianceRebalanceFeeder:
		msg = fmt.Sprintf("<b>[Alliance Hub]</b> Rebalance `ualliance` delegations successfully %s", url)
	case types.AllianceOracleFeeder:
		msg = fmt.Sprintf("<b>[Alliance Oracle]</b> Feeded oracle with data successfully %s", url)
	}

	return a.telegramProvider.SendLog(msg)
}
