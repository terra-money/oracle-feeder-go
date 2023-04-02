package internal

import (
	"fmt"
	"time"

	"github.com/terra-money/oracle-feeder-go/internal/restful"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
	"golang.org/x/exp/maps"
)

type RESTfulProvider struct {
	priceBySymbol map[string]internal_types.PriceBySymbol
}

func NewRESTfulProvider(exchange string, symbols []string, interval int, timeout int, stopCh <-chan struct{}) (*RESTfulProvider, error) {
	client, err := restful.NewRESTfulClient(exchange, symbols)
	if err != nil {
		return nil, err
	}

	provider := &RESTfulProvider{
		priceBySymbol: make(map[string]internal_types.PriceBySymbol),
	}

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		prices, err := client.FetchAndParse(symbols, timeout)
		if err != nil {
			return
		}
		maps.Copy(provider.priceBySymbol, prices)
		for {
			select {
			case <-stopCh:
				ticker.Stop()
				return
			case <-ticker.C:
				prices, err := client.FetchAndParse(symbols, timeout)
				if err == nil {
					maps.Copy(provider.priceBySymbol, prices)
				}
			}
		}
	}()

	return provider, nil
}

func (p *RESTfulProvider) GetPrices() map[string]types.PriceByPair {
	result := make(map[string]types.PriceByPair)
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
