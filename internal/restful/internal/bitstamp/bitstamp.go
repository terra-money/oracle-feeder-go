package bitstamp

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
	exchange   = "bitstamp"
	baseUrl    = "https://www.bitstamp.net/api/v2/ohlc/"
	numWorkers = 16
)

type BitstampClient struct{}

func NewBitstampClient() *BitstampClient {
	return &BitstampClient{}
}

func (p *BitstampClient) FetchAndParse(symbols []string, timeout int) (map[string]internal_types.PriceBySymbol, error) {
	prices := make(map[string]internal_types.PriceBySymbol)

	symbolCh := make(chan string)
	go func() {
		for _, symbol := range symbols {
			symbolCh <- symbol
		}
		close(symbolCh)
	}()

	priceCh := make(chan *internal_types.PriceBySymbol)
	go func() {
		for price := range priceCh {
			prices[price.Symbol] = *price
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			httpWorker(timeout, symbolCh, priceCh)
			wg.Done()
		}()
	}
	wg.Wait()
	return prices, nil
}

func httpWorker(timeout int, symbolCh <-chan string, outCh chan<- *internal_types.PriceBySymbol) {
	for symbol := range symbolCh {
		price, err := fetchSymbol(symbol, timeout)
		if err != nil {
			log.Printf("fetchSymbol(%s) failed: %v", symbol, err)
		}
		outCh <- price
	}
}

// See https://www.bitstamp.net/api/#ohlc_data
type OHLCVResponse struct {
	Data struct {
		OHLC []struct {
			Open      string `json:"open"`
			High      string `json:"high"`
			Low       string `json:"low"`
			Close     string `json:"close"`
			Volume    string `json:"volume"`
			Timestamp string `json:"timestamp"`
		} `json:"ohlc"`
		Pair string `json:"pair"`
	} `json:"data"`
}

// API doc: https://www.bitstamp.net/api/#ohlc_data
func fetchSymbol(symbol string, timeout int) (*internal_types.PriceBySymbol, error) {
	url := fmt.Sprintf("%s/%s/?step=60&limit=1", baseUrl, symbol)
	log.Println(url)
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
	var ohlcResp OHLCVResponse
	err = json.Unmarshal(body, &ohlcResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s", string(body))
	}
	if len(ohlcResp.Data.OHLC) == 0 {
		return nil, fmt.Errorf("no ohlc: %s", string(body))
	}
	ohlcv := ohlcResp.Data.OHLC[0]
	close, err := strconv.ParseFloat(ohlcv.Close, 64)
	if err != nil {
		return nil, err
	}
	open, err := strconv.ParseFloat(ohlcv.Open, 64)
	if err != nil {
		return nil, err
	}
	timestamp, err := strconv.ParseInt(ohlcv.Timestamp, 10, 64)
	if err != nil {
		return nil, err
	}
	price := &internal_types.PriceBySymbol{
		Exchange:  exchange,
		Symbol:    symbol,
		Base:      base,
		Quote:     quote,
		Price:     (open + close) / 2.0,
		Timestamp: uint64(timestamp),
	}
	return price, nil
}
