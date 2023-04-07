package bitstamp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/terra-money/oracle-feeder-go/internal/parser"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	exchange = "bitstamp"
	baseUrl  = "https://www.bitstamp.net/api/v2/ohlc/"
)

type BitstampClient struct{}

func NewBitstampClient() *BitstampClient {
	return &BitstampClient{}
}

// Doc: https://www.bitstamp.net/api/
// OHLC data api only support single currency pair, need fetch one by one.
func (p *BitstampClient) FetchAndParse(symbols []string, timeout int) (map[string]internal_types.PriceBySymbol, error) {
	prices := make(map[string]internal_types.PriceBySymbol)
	for _, symbol := range symbols {
		price, err := fetchSymbol(symbol, timeout)
		if err != nil {
			log.Printf("fetch symbol: %s error\n", symbol)
			continue
		}
		prices[symbol] = *price
	}

	return prices, nil
}

func fetchSymbol(symbol string, timeout int) (*internal_types.PriceBySymbol, error) {
	url := fmt.Sprintf("%s/%s/?step=60&limit=1", baseUrl, symbol)
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	base, quote, err := parser.ParseSymbol(exchange, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to parse symbol %s", symbol)
	}
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
	data, ok := jsonObj["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no data: %s", string(body))
	}
	ohlc, ok := data["ohlc"].([]interface{})
	if !ok || len(ohlc) == 0 {
		return nil, fmt.Errorf("no ohlc: %s", string(body))
	}
	candle := ohlc[0].(map[string]interface{})
	close, err := strconv.ParseFloat(candle["close"].(string), 64)
	if err != nil {
		return nil, err
	}
	open, err := strconv.ParseFloat(candle["open"].(string), 64)
	if err != nil {
		return nil, err
	}
	// high, err := strconv.ParseFloat(candle["high"].(string), 64)
	// if err != nil {
	// 	return nil, err
	// }
	// low, err := strconv.ParseFloat(candle["low"].(string), 64)
	// if err != nil {
	// 	return nil, err
	// }
	baseVolume, err := strconv.ParseFloat(candle["volume"].(string), 64)
	if err != nil {
		return nil, err
	}
	timestamp, err := strconv.ParseInt(candle["timestamp"].(string), 10, 64)
	if err != nil {
		return nil, err
	}
	quoteVolume := 0.0
	vwap := 0.0
	if baseVolume == 0.0 || quoteVolume == 0.0 {
		vwap = (open + close) / 2.0
	} else {
		vwap = quoteVolume / baseVolume
	}
	price := &internal_types.PriceBySymbol{
		Exchange:  exchange,
		Symbol:    symbol,
		Base:      base,
		Quote:     quote,
		Price:     vwap,
		Timestamp: uint64(timestamp),
	}
	return price, nil
}
