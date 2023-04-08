package internal

import (
	"fmt"
	"sync"
	"time"

	"github.com/terra-money/oracle-feeder-go/internal/restful"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
	"golang.org/x/exp/maps"
)

type RESTfulProvider struct {
	priceBySymbol map[string]internal_types.PriceBySymbol
	mu            *sync.Mutex
}

func NewRESTfulProvider(exchange string, symbols []string, interval int, timeout int, stopCh <-chan struct{}) (*RESTfulProvider, error) {
	client, err := restful.NewRESTfulClient(exchange, symbols)
	if err != nil {
		return nil, err
	}

	mu := sync.Mutex{}
	provider := &RESTfulProvider{
		priceBySymbol: make(map[string]internal_types.PriceBySymbol),
		mu:            &mu,
	}

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		prices, err := client.FetchAndParse(symbols, timeout)
		if err == nil {
			mu.Lock()
			maps.Copy(provider.priceBySymbol, prices)
			mu.Unlock()
		}
		for {
			select {
			case <-stopCh:
				ticker.Stop()
				return
			case <-ticker.C:
				prices, err := client.FetchAndParse(symbols, timeout)
				if err == nil {
					mu.Lock()
					maps.Copy(provider.priceBySymbol, prices)
					mu.Unlock()
				}
			}
		}
	}()

	return provider, nil
}

func (p *RESTfulProvider) GetPrices() map[string]types.PriceByPair {
	result := make(map[string]types.PriceByPair)
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, price := range p.priceBySymbol {
		pair := fmt.Sprintf("%s/%s", price.Base, price.Quote)
		result[pair] = types.PriceByPair{
			Base:      price.Base,
			Quote:     price.Quote,
			Price:     price.Price,
			Timestamp: price.Timestamp,
		}
	}
	return result
}
