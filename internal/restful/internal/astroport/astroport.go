package astroport

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/terra-money/oracle-feeder-go/internal/parser"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
)

type AstroportClient struct {
	url     string
	amount  int64
	chainId string
}

func NewAstroportClient() *AstroportClient {
	return &AstroportClient{
		url:     "https://develop-multichain-api.astroport.fi/router/v2/routes",
		amount:  1000000,
		chainId: "phoenix-1",
	}
}

func (p *AstroportClient) FetchAndParse(symbols []string, timeout int) (map[string]internal_types.PriceBySymbol, error) {
	astroData, err := p.fetchPrices(symbols, timeout)
	if err != nil {
		return nil, err
	}
	res := p.parseAmounts(astroData)
	fmt.Print(res)
	return res, nil
}

func (p *AstroportClient) fetchPrices(symbols []string, timeout int) (map[string]AstroportData, error) {
	astroData := make(map[string]AstroportData)
	for _, symbol := range symbols {
		symbolSplit := strings.Split(symbol, "-")
		data, err := p.queryData(symbolSplit[0], symbolSplit[1])
		if err != nil {
			return nil, err
		}

		astroData[symbol] = data[0]
		if err != nil {
			return nil, err
		}

	}
	return astroData, nil
}

func (p *AstroportClient) queryData(start, end string) (res []AstroportData, err error) {
	urlParams := fmt.Sprintf("?start=%s&end=%s&amount=%d&chainId=%s", start, end, p.amount, p.chainId)

	// Send GET request
	resp, err := http.Get(p.url + urlParams)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response into struct
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	// Access parsed data
	return res, nil
}

func (p *AstroportClient) parseAmounts(res map[string]AstroportData) map[string]internal_types.PriceBySymbol {
	prices := make(map[string]internal_types.PriceBySymbol)
	now := uint64(time.Now().UnixMilli())
	for symbol, value := range res {
		base, quote, err := parser.ParseSymbol("astroport", symbol)
		if err == nil {
			prices[symbol] = internal_types.PriceBySymbol{
				Exchange:  "astroport",
				Symbol:    symbol,
				Base:      base,
				Quote:     quote,
				Price:     float64(value.AmountOut) / math.Pow10(6), // 6 decimals
				Timestamp: now,
			}
		} else {
			log.Printf("%v", err)
		}
	}
	return prices
}
