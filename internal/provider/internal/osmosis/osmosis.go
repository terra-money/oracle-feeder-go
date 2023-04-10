package osmosis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/terra-money/oracle-feeder-go/configs"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
	"golang.org/x/exp/maps"
)

type OsmosisEndpoint struct {
	url  string
	used bool
}

type OsmosisProvider struct {
	priceBySymbol map[string]internal_types.PriceBySymbol
	config        *configs.ProviderConfig
	mu            *sync.Mutex
}

var endpoints []OsmosisEndpoint

var whiteListPoolIds = map[string]string{
	"ATOM/USDC": "1",
	"AKT/USDC":  "3",
	// "CRO/USDC": "9",      // DOUBLE CHECK
	"JUNO/USDC": "497",
	// "USTC/USDC": "560",   // DOUBLE CHECK
	"SCRT/USDC":  "584",
	"STARS/USDC": "604",
	// "DAI/USDC": "674",    // DOUBLE CHECK
	"OSMO/USDC": "678",
	// "EVMOS/USDC": "722",  // DOUBLE CHECK
	"INJ/USDC":  "725",
	"LUNA/USDC": "726",
	"KAVA/USDC": "730",
	"LINK/USDC": "731",
	// "MKR/USDC": "733",    // DOUBLE CHECK
	// "DOT/USDC": "773",    // DOUBLE CHECK
	"LUNC/USDC": "800",
}
var idToSymbols = make(map[string]string)

func NewOsmosisProvider(config *configs.ProviderConfig, stopCh <-chan struct{}) (*OsmosisProvider, error) {
	mu := sync.Mutex{}
	provider := &OsmosisProvider{
		priceBySymbol: make(map[string]internal_types.PriceBySymbol),
		config:        config,
		mu:            &mu,
	}

	endpoints = append(endpoints, OsmosisEndpoint{
		"https://osmosis-api.polkachu.com/osmosis/gamm/v1beta1/pools?pagination.limit=801",
		false,
	})
	endpoints = append(endpoints, OsmosisEndpoint{
		"https://lcd.osmosis.zone/osmosis/gamm/v1beta1/pools?pagination.limit=801",
		false,
	})
	for k, v := range whiteListPoolIds {
		idToSymbols[v] = k
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

func rotateUrl() (string, error) {
	if len(endpoints) == 0 {
		return "", fmt.Errorf("No endpoints")
	}
	for i := range endpoints {
		if !endpoints[i].used {
			endpoints[i].used = true
			return endpoints[i].url, nil
		}
	}
	for i := range endpoints {
		endpoints[i].used = false
	}
	endpoints[0].used = true
	return endpoints[0].url, nil
}

func (p *OsmosisProvider) GetPrices() map[string]types.PriceByPair {
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

func (p *OsmosisProvider) fetchAndParse() {
	msg, err := fetchPrices(p.config.Symbols)
	if err != nil {
		log.Printf("%v", err)
	} else {
		prices, err := parseJSON(msg)
		if err != nil {
			log.Printf("%v", err)
		} else {
			p.mu.Lock()
			maps.Copy(p.priceBySymbol, prices)
			p.mu.Unlock()
		}
	}
}

func fetchPrices(symbols []string) ([]interface{}, error) {
	url, err := rotateUrl()
	// fmt.Printf("url: %s\n", url)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Second * 15}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jsonObj := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonObj)
	if err != nil {
		log.Printf("parse response error: %v\n", string(body))
		return nil, err
	}
	data, ok := jsonObj["pools"].([]interface{})
	if !ok {
		log.Printf("no pools: %v\n", string(body))
		return nil, err
	}
	return data, nil
}

func parseJSON(msg []interface{}) (map[string]internal_types.PriceBySymbol, error) {
	prices := make(map[string]internal_types.PriceBySymbol)
	now := time.Now()

	osmoPrice := 0.0
	osmoPair := "OSMO/USDC"

	for _, value := range msg {
		item := value.(map[string]interface{})
		poolId := item["id"].(string)
		symbol, ok := idToSymbols[poolId]
		if !ok {
			continue
		}
		base, quote, err := parser.ParseSymbol("osmosis", symbol)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		assets, ok := item["pool_assets"].([]interface{})
		if !ok || len(assets) != 2 {
			log.Printf("invalid pool_assets: %v\n", item)
			continue
		}
		first := assets[0].(map[string]interface{})["token"].(map[string]interface{})
		firstAmount, err := strconv.ParseUint(first["amount"].(string), 10, 64)
		if err != nil {
			continue
		}
		second := assets[1].(map[string]interface{})["token"].(map[string]interface{})
		// secondDenom := second["denom"].(string)
		secondAmount, err := strconv.ParseUint(second["amount"].(string), 10, 64)
		if err != nil {
			continue
		}
		price := 0.0
		// log.Printf("secondDenom: %s base: %s %v quote: %s %v\n", secondDenom, base, firstAmount, quote, secondAmount)
		if firstAmount > 0 && secondAmount > 0 {
			// if secondDenom == "uosmo" {
			if symbol == osmoPair {
				price = float64(firstAmount) / float64(secondAmount)
			} else {
				price = float64(secondAmount) / float64(firstAmount)
			}
		}
		if symbol == osmoPair {
			osmoPrice = price
		}
		prices[symbol] = internal_types.PriceBySymbol{
			Exchange:  "osmosis",
			Symbol:    symbol,
			Base:      base,
			Quote:     quote,
			Price:     price,
			Timestamp: uint64(now.UnixMilli()),
		}
	}
	if osmoPrice == 0 {
		return nil, fmt.Errorf("no osmo price")
	}
	for _, pairPrice := range prices {
		if pairPrice.Symbol != osmoPair {
			pairPrice.Price = pairPrice.Price * osmoPrice
			prices[pairPrice.Symbol] = pairPrice
		}
	}
	// fmt.Printf("osmoPrice: %v prices: %v\n", osmoPrice, prices)
	return prices, nil
}
