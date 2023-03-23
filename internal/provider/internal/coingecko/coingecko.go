package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/terra-money/oracle-feeder-go/configs"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
	"golang.org/x/exp/maps"
)

const (
	coingeckoEndpoint string = "https://api.coingecko.com/api/v3/simple/price"
)

type CoingeckoProvider struct {
	priceBySymbol map[string]internal_types.PriceBySymbol
	config        *configs.ProviderConfig
}

func NewCoingeckoProvider(config *configs.ProviderConfig, stopCh <-chan struct{}) (*CoingeckoProvider, error) {
	provider := &CoingeckoProvider{
		priceBySymbol: make(map[string]internal_types.PriceBySymbol),
		config:        config,
	}

	go func() {
		ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
		provider.fetchAndParse()
		for {
			select {
			case <-stopCh:
				ticker.Stop()
				return
			case <-ticker.C:
				provider.fetchAndParse()
			}
		}
	}()

	return provider, nil
}

func (p *CoingeckoProvider) GetPrices() map[string]types.PriceByPair {
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

func (p *CoingeckoProvider) fetchAndParse() {
	msg, err := fetchPrices(p.config.Symbols)
	if err != nil {
		log.Printf("%v", err)
	} else {
		prices := parseJSON(msg)
		maps.Copy(p.priceBySymbol, prices)
	}
}

func fetchPrices(symbols []string) (map[string]map[string]float64, error) {
	params := url.Values{}
	params.Add("vs_currencies", "usd")
	params.Add("precision", "18")
	params.Add("ids", strings.Join(symbols, ","))
	url := fmt.Sprintf("%s?%s", coingeckoEndpoint, params.Encode())
	client := &http.Client{Timeout: time.Second * 15}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jsonObj := make(map[string]map[string]float64)
	err = json.Unmarshal(body, &jsonObj)
	if err != nil {
		return nil, err
	}
	return jsonObj, nil
}

func parseJSON(msg map[string]map[string]float64) map[string]internal_types.PriceBySymbol {
	prices := make(map[string]internal_types.PriceBySymbol)
	now := time.Now()
	for symbol, value := range msg {
		base, quote, err := parser.ParseSymbol("coingecko", symbol)
		if err == nil {
			prices[symbol] = internal_types.PriceBySymbol{
				Exchange:  "coingecko",
				Symbol:    symbol,
				Base:      base,
				Quote:     quote,
				Price:     value["usd"],
				Timestamp: uint64(now.UnixMilli()),
			}
		} else {
			log.Printf("%v", err)
		}
	}
	return prices
}
