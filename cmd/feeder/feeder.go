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
)

func main() {
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

	ctx := context.Background()
	alliancesQuerierProvider := alliance_provider.NewAlliancesQuerierProvider()

	for attempt := 1; attempt <= retries; attempt++ {
		_, err := alliancesQuerierProvider.QueryAndSubmitOnChain(ctx)

		if err == nil {
			// Code executed successfully
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
