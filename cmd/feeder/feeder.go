package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/terra-money/oracle-feeder-go/internal/provider"
	"github.com/terra-money/oracle-feeder-go/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	ctx := context.Background()
	provider := provider.NewTransactionsProvider()

	res, err := requestAlliancesData()
	if err != nil {
		log.Fatal("ERROR requesting alliances data", err)
	}
	msg, err := provider.ParseAlliancesTransaction(res)
	if err != nil {
		log.Fatal("ERROR parsing alliances data", err)
	}
	txHash, err := provider.SubmitAlliancesTransaction(ctx, []sdk.Msg{msg})
	if err != nil {
		log.Fatal("ERROR submittin alliances data on chain ", err)
	}

	fmt.Printf("Transaction Submitted successfully txHash: %d \n", txHash)
}

func requestAlliancesData() (res []types.ProtocolInfo, err error) {
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
