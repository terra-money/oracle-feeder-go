package kucoin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	exchangeName string = "kucoin"
)

type WebsocketClient struct{}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{}
}

func (wc *WebsocketClient) ConnectAndSubscribe(symbols []string) (*websocket.Conn, error) {
	wsToken, err := fetchWebsocketToken()
	if err != nil {
		return nil, err
	}
	wsUrl := fmt.Sprintf("%s?token=%s&connectId=terra-price-server", wsToken.endpoint, wsToken.token)
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return nil, err
	}

	resp := make(map[string]interface{})
	if err := conn.ReadJSON(&resp); err != nil {
		return nil, err
	}
	if typ, ok := resp["type"].(string); ok && typ != "welcome" {
		bytes, _ := json.Marshal(resp)
		return nil, fmt.Errorf("%s", string(bytes))
	}

	// send ping per 30 seconds
	// see https://docs.kucoin.com/#ping
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			pingMsg := make(map[string]any)
			pingMsg["type"] = "ping"
			pingMsg["id"] = "terra-price-server"
			err = conn.WriteJSON(pingMsg)
			if err != nil {
				log.Printf("%v", err)
			}
		}
	}()

	commands := generateCommands(symbols)
	for _, command := range commands {
		if err := conn.WriteJSON(&command); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

// Candlestick websocket message.
//
// Message format: https://docs.kucoin.com/#klines
type RawCandlestickMsg struct {
	Type    string `json:"type"`
	Topic   string `json:"topic"`
	Subject string `json:"subject"`
	Data    struct {
		Symbol  string   `json:"symbol"`
		Candles []string `json:"candles"`
		TimeUs  uint64   `json:"time"`
	} `json:"data"`
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://docs.kucoin.com/#klines
//
// Kucoin allows to subscribe 100 symbols per time, otherwise you'll get an error
// "exceed max subscription count limitation of 100 per time"
//
// For example:
// {"id":"terra-price-server","type":"subscribe","topic":"/market/candles:BTC-USDT_1min,BTC-USDT_1week","privateChannel":false,"response":true}
func generateCommand(symbols []string) (map[string]interface{}, error) {
	if len(symbols) > 100 {
		return nil, fmt.Errorf("exceeds 100 symbols")
	}
	var topics []string
	for _, symbol := range symbols {
		topics = append(topics, fmt.Sprintf("%s_1min", symbol))
	}
	topic := fmt.Sprintf("/market/candles:%s", strings.Join(topics, ","))
	return map[string]interface{}{
		"id":             "terra-price-server",
		"type":           "subscribe",
		"topic":          topic,
		"response":       false,
		"privateChannel": false,
	}, nil
}

func generateCommands(symbols []string) []map[string]interface{} {
	var commands []map[string]interface{}
	groupSize := 100
	n := len(symbols)
	for i := 0; i < n; i += groupSize {
		j := i + groupSize
		if j > n {
			j = n
		}
		group := symbols[i:j]
		command, err := generateCommand(group)
		if err == nil {
			commands = append(commands, command)
		}
	}
	return commands
}

func (wc *WebsocketClient) ParseCandlestickMsg(rawMsg []byte) (*types.CandlestickMsg, error) {
	var msg RawCandlestickMsg
	json.Unmarshal(rawMsg, &msg)

	candles := msg.Data.Candles
	if len(candles) != 7 {
		return nil, fmt.Errorf("invalid candles %s", string(rawMsg))
	}

	symbol := msg.Data.Symbol
	base, quote, err := parser.ParseSymbol(exchangeName, symbol)
	if err != nil {
		return nil, err
	}

	open, err := strconv.ParseFloat(candles[1], 64)
	if err != nil {
		return nil, err
	}
	close, err := strconv.ParseFloat(candles[2], 64)
	if err != nil {
		return nil, err
	}
	high, err := strconv.ParseFloat(candles[3], 64)
	if err != nil {
		return nil, err
	}
	low, err := strconv.ParseFloat(candles[4], 64)
	if err != nil {
		return nil, err
	}
	baseVolume, err := strconv.ParseFloat(candles[5], 64)
	if err != nil {
		return nil, err
	}
	quoteVolume, err := strconv.ParseFloat(candles[6], 64)
	if err != nil {
		return nil, err
	}
	vwap := 0.0
	if baseVolume == 0.0 || quoteVolume == 0.0 {
		vwap = (open + close) / 2.0
	} else {
		vwap = quoteVolume / baseVolume
	}

	return &types.CandlestickMsg{
		Exchange:  exchangeName,
		Symbol:    symbol,
		Base:      base,
		Quote:     quote,
		Timestamp: msg.Data.TimeUs / 1e3,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    baseVolume,
		Vwap:      vwap,
	}, nil
}

type websocketToken struct {
	token    string
	endpoint string
}

// see https://docs.kucoin.com/#apply-connect-token
func fetchWebsocketToken() (*websocketToken, error) {
	url := "https://openapi-v2.kucoin.com/api/v1/bullet-public"
	client := &http.Client{Timeout: time.Second * 15}
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	jsonObj := make(map[string]any)
	err = json.Unmarshal(body, &jsonObj)
	if err != nil {
		return nil, err
	}

	code := jsonObj["code"].(string)
	if code != "200000" {
		return nil, fmt.Errorf("%s", string(body))
	}

	data := jsonObj["data"].(map[string]any)
	token := data["token"].(string)
	instanceServers := data["instanceServers"].([]any)
	endpoint := instanceServers[0].(map[string]any)["endpoint"].(string)
	return &websocketToken{
		token:    token,
		endpoint: endpoint,
	}, nil
}
