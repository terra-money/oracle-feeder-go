package fer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	exchange = "fer"
	baseUrl  = "https://api.fer.ee/latest"
)

type FerClient struct{}

func NewFerClient() *FerClient {
	return &FerClient{}
}

func (p *FerClient) FetchAndParse(symbols []string, timeout int) (map[string]internal_types.PriceBySymbol, error) {
	var baseCurrencies []string
	for _, symbol := range symbols {
		items := strings.Split(symbol, "/")
		baseCurrencies = append(baseCurrencies, items[0])
	}
	url := fmt.Sprintf("%s?base=USD&to=%s", baseUrl, strings.Join(baseCurrencies, ","))
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	log.Println(url)
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
		return nil, fmt.Errorf("failed to unmarshal %s", string(body))
	}
	rates, ok := jsonObj["rates"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no rates: %s", string(body))
	}
	prices := make(map[string]internal_types.PriceBySymbol)
	now := uint64(time.Now().UnixMilli())
	quote := "USD"
	for base, v := range rates {
		symbol := fmt.Sprintf("%s/%s", base, quote)
		price := v.(float64)
		if price > 0 {
			prices[symbol] = internal_types.PriceBySymbol{
				Exchange:  exchange,
				Symbol:    symbol,
				Base:      base,
				Quote:     quote,
				Price:     1.0 / price,
				Timestamp: now,
			}
		}
	}
	return prices, nil
}
