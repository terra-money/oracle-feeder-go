package internal

import (
	"fmt"

	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/internal/websocket"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
)

type WebsocketProvider struct {
	priceBySymbol map[string]internal_types.PriceBySymbol
}

func NewWebsocketProvider(exchange string, symbols []string, stopCh <-chan struct{}) (*WebsocketProvider, error) {
	candlestickCh, err := websocket.SubscribeCandlestick(exchange, symbols, stopCh)
	if err != nil {
		return nil, err
	}
	provider := &WebsocketProvider{
		priceBySymbol: make(map[string]internal_types.PriceBySymbol),
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
			provider.priceBySymbol[msg.Symbol] = price
		}
	}()
	return provider, nil
}

func (p *WebsocketProvider) GetPrices() map[string]types.PriceByPair {
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
