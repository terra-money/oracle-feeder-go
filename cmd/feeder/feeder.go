package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

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

	// Create a channel that will receive a value at the specified interval
	ticker := time.Tick(5 * time.Second)

	// Start a goroutine to perform the action
	go func() {
		for range ticker {
			res, err := requestAlliancesData()
			if err != nil {
				panic(err)
			}
			msg, err := provider.ParseAlliancesTransaction(res)
			if err != nil {
				panic(err)
			}
			txHash, err := provider.SubmitAlliancesTransaction(ctx, []sdk.Msg{msg})
			if err != nil {
				panic(err)
			}

			fmt.Printf("Transaction Submitted successfully txHash: %d \n", txHash)
		}
	}()

	// Wait for user input to exit the program
	fmt.Println("Press Enter to stop...")
	fmt.Scanln()
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
