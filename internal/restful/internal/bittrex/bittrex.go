package bittrex

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/terra-money/oracle-feeder-go/internal/parser"
	internal_types "github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	exchange = "bittrex"
	// API doc: https://bittrex.github.io/api/v3
	baseUrl    = "https://api.bittrex.com/v3"
	numWorkers = 16
)

type BittrexClient struct{}

func NewBittrexClient() *BittrexClient {
	return &BittrexClient{}
}

// Candle api only support single currency pair, need fetch one by one.
func (p *BittrexClient) FetchAndParse(symbols []string, timeout int) (map[string]internal_types.PriceBySymbol, error) {
	prices := make(map[string]internal_types.PriceBySymbol)
	mu := sync.Mutex{}

	symbolCh := make(chan string)
	go func() {
		for _, symbol := range symbols {
			symbolCh <- symbol
		}
		close(symbolCh)
	}()

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			httpWorker(timeout, symbolCh, &mu, prices)
			wg.Done()
		}()
	}
	wg.Wait()
	return prices, nil
}

func httpWorker(timeout int, symbolCh <-chan string, mu *sync.Mutex, prices map[string]internal_types.PriceBySymbol) {
	for symbol := range symbolCh {
		price, err := fetchCandle(symbol, timeout)
		if err != nil {
			log.Printf("fetchCandle(%s) failed: %v", symbol, err)
		}
		mu.Lock()
		prices[price.Symbol] = *price
		mu.Unlock()
	}
}

// API doc https://bittrex.github.io/api/v3#operation--markets--marketSymbol--candles--candleType---candleInterval--recent-get
func fetchCandle(symbol string, timeout int) (*internal_types.PriceBySymbol, error) {
	url := fmt.Sprintf("%s/markets/%s/candles/trade/MINUTE_1/recent", baseUrl, symbol)
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
	var jsonObj []interface{}
	err = json.Unmarshal(body, &jsonObj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s", string(body))
	}
	if len(jsonObj) == 0 {
		return nil, fmt.Errorf("empty array %s", string(body))
	}

	candle, ok := jsonObj[len(jsonObj)-1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no data: %s", string(body))
	}
	close, err := strconv.ParseFloat(candle["close"].(string), 64)
	if err != nil {
		return nil, err
	}
	open, err := strconv.ParseFloat(candle["open"].(string), 64)
	if err != nil {
		return nil, err
	}

	startsAt, err := time.Parse(time.RFC3339, candle["startsAt"].(string))
	if err != nil {
		return nil, err
	}
	endsAt := startsAt.Add(time.Minute)

	baseVolume, err := strconv.ParseFloat(candle["volume"].(string), 64)
	if err != nil {
		return nil, err
	}
	quoteVolume, err := strconv.ParseFloat(candle["quoteVolume"].(string), 64)
	if err != nil {
		return nil, err
	}
	vwap := 0.0
	if baseVolume == 0.0 || quoteVolume == 0.0 {
		vwap = (open + close) / 2.0
	} else {
		vwap = quoteVolume / baseVolume
	}

	return &internal_types.PriceBySymbol{
		Exchange:  exchange,
		Symbol:    symbol,
		Base:      base,
		Quote:     quote,
		Price:     vwap,
		Timestamp: uint64(endsAt.UnixMilli()),
	}, nil
}
