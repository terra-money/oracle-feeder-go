package internal

import (
	"fmt"
	"sync"

	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/internal/websocket"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
)

type WebsocketProvider struct {
	priceBySymbol map[string]internal_types.PriceBySymbol
	mu            *sync.Mutex
}

func NewWebsocketProvider(exchange string, symbols []string, stopCh <-chan struct{}) (*WebsocketProvider, error) {
	candlestickCh, err := websocket.SubscribeCandlestick(exchange, symbols, stopCh)
	if err != nil {
		return nil, err
	}

	mu := sync.Mutex{}
	provider := &WebsocketProvider{
		priceBySymbol: make(map[string]internal_types.PriceBySymbol),
		mu:            &mu,
	}

	go func() {
		for msg := range candlestickCh {
			price := internal_types.PriceBySymbol{
				Exchange:  msg.Exchange,
				Symbol:    msg.Symbol,
				Base:      msg.Base,
				Quote:     msg.Quote,
				Price:     msg.Vwap,
				Timestamp: msg.Timestamp,
			}
			mu.Lock()
			provider.priceBySymbol[msg.Symbol] = price
			mu.Unlock()
		}
	}()
	return provider, nil
}

func (p *WebsocketProvider) GetPrices() map[string]types.PriceByPair {
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
