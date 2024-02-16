package osmosis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/oracle-feeder-go/config"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/pkg/types"
)

type OsmosisEndpoint struct {
	url  string
	used bool
}

type OsmosisProvider struct {
	priceBySymbol map[string]internal_types.PriceBySymbol
	config        *config.ProviderConfig
	mu            *sync.Mutex
}

var endpoints = []OsmosisEndpoint{{
	url:  "https://osmosis-api.polkachu.com/osmosis/gamm/v1beta1/pools/${POOL_ID}",
	used: false,
}, {
	url:  "https://osmosis-api.polkachu.com/osmosis/gamm/v1beta1/pools/${POOL_ID}",
	used: false,
}, {
	url:  "https://lcd-osmosis.tfl.foundation/osmosis/gamm/v1beta1/pools/${POOL_ID}",
	used: false,
}}

var whiteListPoolIds = map[string]string{
	"ATOM/OSMO":  "1",
	"AKT/OSMO":   "3",
	"JUNO/OSMO":  "497",
	"SCRT/OSMO":  "584",
	"STARS/OSMO": "604",
	"USDC/OSMO":  "678",
	"INJ/OSMO":   "725",
	"LUNA/OSMO":  "726",
	"KAVA/OSMO":  "730",
	"LINK/OSMO":  "731",
	"LUNC/OSMO":  "800",
	"ASH/USDC":   "1360",
	"OSMO/USDC":  "1464",
}
var idsBySymbol = make(map[string]string)

func NewOsmosisProvider(config *config.ProviderConfig, stopCh <-chan struct{}) (*OsmosisProvider, error) {
	mu := sync.Mutex{}
	provider := &OsmosisProvider{
		priceBySymbol: make(map[string]internal_types.PriceBySymbol),
		config:        config,
		mu:            &mu,
	}

	for k, v := range whiteListPoolIds {
		idsBySymbol[v] = k
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
	for id, symbol := range idsBySymbol {
		res, err := fetchPrice(id)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		var generic internal_types.GenericPoolResponse
		if err := json.Unmarshal(res, &generic); err != nil {
			fmt.Println("Error:", err)
			continue
		}
		price, err := p.parsePrice(generic, res)
		if err != nil {
			continue
		}

		p.mu.Lock()
		p.priceBySymbol[symbol] = internal_types.PriceBySymbol{
			Symbol:    symbol,
			Price:     price,
			Base:      strings.Split(symbol, "/")[0],
			Quote:     strings.Split(symbol, "/")[1],
			Timestamp: uint64(time.Now().Unix()),
		}
		p.mu.Unlock()
	}
}

func (*OsmosisProvider) parsePrice(generic internal_types.GenericPoolResponse, res []byte) (float64, error) {
	switch generic.Pool.Type {
	case "/osmosis.concentratedliquidity.v1beta1.Pool":
		var pool internal_types.OsmosisPoolResponse
		if err := json.Unmarshal(res, &pool); err != nil {
			return 0, err
		}

		// get the first 18 positons of the price to avoid overflow
		price := pool.Pool.CurrentSqrtPrice
		if len(price) >= 18 {
			price = pool.Pool.CurrentSqrtPrice[:18]
		}
		parsedPrice, err := sdktypes.NewDecFromStr(price)
		if err != nil {
			return 0, err
		}
		return parsedPrice.Power(2).Float64()
	case "/osmosis.gamm.v1beta1.Pool":
		var pool internal_types.OsmosisGammPoolResponse
		if err := json.Unmarshal(res, &pool); err != nil {
			return 0, err
		}

		// get the first 18 positons of the price to avoid overflow
		firstTokenPrice := pool.Pool.PoolAssets[0].Token.Amount
		if len(firstTokenPrice) >= 18 {
			firstTokenPrice = firstTokenPrice[:18]
		}
		parsedFirstTokenPrice, err := sdktypes.NewDecFromStr(firstTokenPrice)
		if err != nil {
			return 0, err
		}
		secondTokenPrice := pool.Pool.PoolAssets[1].Token.Amount
		if len(secondTokenPrice) >= 18 {
			secondTokenPrice = secondTokenPrice[:18]
		}
		parsedSecondTokenPrice, err := sdktypes.NewDecFromStr(secondTokenPrice)
		if err != nil {
			return 0, err
		}
		return parsedSecondTokenPrice.Quo(parsedFirstTokenPrice).Float64()
	default:
		return 0, fmt.Errorf("unknown pool type: %s", generic.Pool.Type)
	}
}

func fetchPrice(poolId string) (res []byte, err error) {
	url, err := rotateUrl()
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Second * 15}
	resp, err := client.Get(strings.Replace(url, "${POOL_ID}", poolId, 1))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func rotateUrl() (string, error) {
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
