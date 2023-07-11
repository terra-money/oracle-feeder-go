package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	alliance_provider "github.com/terra-money/oracle-feeder-go/internal/provider/alliance"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

func main() {
	// Load the environment variables
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file:", err)
	}
	retries := 3
	if feederRetries := os.Getenv("FEEDER_RETRIES"); feederRetries != "" {
		retries, err = strconv.Atoi(feederRetries)
		if err != nil {
			log.Fatal("Error parsing FEEDER_RETRIES:", err)
		}
	}
	// Read the cli arguments
	if len(os.Args) != 2 || os.Args[1] == "" {
		log.Fatal(`Specify the first argument as the feeder type.`)
	}
	feederType, err := types.ParseFeederTypeFromString(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	alliancesQuerierProvider := alliance_provider.NewAlliancesQuerierProvider(feederType)

	for attempt := 1; attempt <= retries; attempt++ {
		txHash, err := alliancesQuerierProvider.SubmitTx(ctx)

		if err == nil {
			fmt.Printf("Transaction Submitted successfully txHash: %s \n", txHash)
			break
		} else {
			// Code execution failed
			fmt.Printf("Attempt %d failed: %v\n", attempt, err)

			if attempt == retries {
				fmt.Println("All attempts failed. Exiting...")
				break
			}

			fmt.Printf("Retrying in 15 seconds...\n")
			time.Sleep(15 * time.Second)
		}
	}
}
