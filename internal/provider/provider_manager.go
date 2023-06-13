package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/terra-money/oracle-feeder-go/config"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
)

type ProviderManager struct {
	config    *config.Config
	providers map[string]Provider
}

func NewProviderManager(config *config.Config, stopCh <-chan struct{}) *ProviderManager {
	providers := make(map[string]Provider)
	for _, exchange := range config.ProviderPriority {
		providerConfig := config.Providers[exchange]
		provider, err := NewProvider(exchange, &providerConfig, stopCh)
		if err != nil {
			panic(err)
		}
		providers[exchange] = provider
	}
	return &ProviderManager{
		config:    config,
		providers: providers,
	}
}

func (m *ProviderManager) GetPrices(ctx context.Context) *types.PricesResponse {
	// exchange -> base -> price
	prices := make(map[string]map[string]types.PriceByPair)
	for exchange, provider := range m.providers {
		prices[exchange] = provider.GetPrices()
	}

	priceByCoin := averagePriceByCoin(averagePriceByPair(prices))

	var pricesOfCoins []types.PriceOfCoin
	now := uint64(time.Now().UnixMilli())
	for coin, price := range priceByCoin {
		pricesOfCoins = append(pricesOfCoins, types.PriceOfCoin{
			Denom:     coin,
			Price:     price,
			Timestamp: now,
		})
	}
	resp := &types.PricesResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Prices:    pricesOfCoins,
	}

	return resp
}

func (m *ProviderManager) GetPrice(ctx context.Context, denom string) *types.PriceResponse {
	r := m.GetPrices(ctx)
	for _, price := range r.Prices {
		if strings.EqualFold(price.Denom, denom) {
			return &types.PriceResponse{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Price:     price,
			}
		}
	}

	return nil
}

// Calculate average price for each pair.
//
// Returns map of pair -> price.
func averagePriceByPair(prices map[string]map[string]types.PriceByPair) map[string]types.PriceByPair {
	pairSum := make(map[string]float64)
	pairCount := make(map[string]float64)
	for _, priceByPair := range prices {
		for _, price := range priceByPair {
			pair := fmt.Sprintf("%s/%s", price.Base, price.Quote)
			pairSum[pair] += price.Price
			pairCount[pair] += 1.0
		}
	}
	averagedPrices := make(map[string]types.PriceByPair)
	for pair, count := range pairCount {
		arr := strings.Split(pair, "/")
		base := arr[0]
		quote := arr[1]
		averagedPrices[pair] = types.PriceByPair{
			Base:      base,
			Quote:     quote,
			Price:     pairSum[pair] / count,
			Timestamp: uint64(time.Now().UnixMilli()),
		}
	}
	return averagedPrices
}

func averagePriceByCoin(prices map[string]types.PriceByPair) map[string]float64 {
	coinSum := make(map[string]float64)
	coinCount := make(map[string]float64)
	for _, priceByPair := range prices {
		if priceByPair.Quote == "USD" {
			coinSum[priceByPair.Base] += priceByPair.Price
			coinCount[priceByPair.Base] += 1.0
		} else {
			quoteUSD := fmt.Sprintf("%s/USD", priceByPair.Quote)
			if quotePrice, ok := prices[quoteUSD]; ok && quotePrice.Price > 0.0 {
				coinSum[priceByPair.Base] += priceByPair.Price * quotePrice.Price
				coinCount[priceByPair.Base] += 1.0
			} else {
				println("line 106 ", quoteUSD)
			}
		}
	}

	priceByCoin := make(map[string]float64)
	for coin, count := range coinCount {
		priceByCoin[coin] = coinSum[coin] / count
	}
	return priceByCoin
}
