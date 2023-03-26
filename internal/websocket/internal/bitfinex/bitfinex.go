package bitfinex

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	websocketUrl string = "wss://api-pub.bitfinex.com/ws/2"
	exchangeName string = "bitfinex"
)

var idToChannels = map[uint64]string{}

type WebsocketClient struct{}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{}
}

func (wc *WebsocketClient) ConnectAndSubscribe(symbols []string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl, nil)
	if err != nil {
		return nil, err
	}

	resp := make(map[string]interface{})
	if err := conn.ReadJSON(&resp); err != nil {
		return nil, err
	}
	event, ok := resp["event"].(string)
	if ok && event == "info" {
		code, ok := resp["code"].(float64)
		if ok {
			return nil, fmt.Errorf("connect error, code: %v %v", code, resp)
		}
		platform := resp["platform"].(map[string]interface{})
		status := uint64(platform["status"].(float64))
		if status != 1 {
			return nil, fmt.Errorf("connect error, code: %v", resp)
		}
	} else {
		return nil, fmt.Errorf("connect error: %v", resp)
	}

	for _, symbol := range symbols {
		command := generateCommand(symbol)
		if err := conn.WriteJSON(&command); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (wc *WebsocketClient) HandleMsg(rawMsg []byte, conn *websocket.Conn) (*types.CandlestickMsg, error) {
	if strings.HasPrefix(string(rawMsg), "[") {
		return parseCandlestickMsg(conn, rawMsg)
	}
	if strings.HasPrefix(string(rawMsg), "{") {
		resp := make(map[string]interface{})
		if err := json.Unmarshal(rawMsg, &resp); err != nil {
			return nil, err
		}
		event, ok := resp["event"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid msg: %s", string(rawMsg))
		}
		if event == "subscribed" {
			channelId := uint64(resp["chanId"].(float64))
			key := resp["key"].(string)
			items := strings.Split(key, ":")
			if len(items) < 3 {
				return nil, fmt.Errorf("invalid subscribed: %s", string(rawMsg))
			}
			idToChannels[channelId] = items[2]
			return nil, nil
		}
		if event == "error" {
			log.Printf("Error msg: %v\n", string(rawMsg))
			return nil, nil
		}
		if event == "pong" {
			return nil, nil
		}
	}
	log.Printf("Unrecognized msg: %v\n", string(rawMsg))
	return nil, nil
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://docs.bitfinex.com/reference/ws-public-candles
//
// For example:
// { "event": "subscribe",  "channel": "candles",  "key": "trade:1m:tBTCUSD" }
func generateCommand(symbol string) map[string]interface{} {
	key := fmt.Sprintf("trade:1m:%s", symbol)
	return map[string]interface{}{
		"event":   "subscribe",
		"channel": "candles",
		"key":     key,
	}
}

// API doc: https://docs.bitfinex.com/reference/ws-public-candles
// For example:
// [190359,"hb"]
// [190359,[1679725500000,27472,27479,27479,27472,0.0734273]]
// [190359,[[1679725500000,27472,27479,27479,27472,0.0734273]]]
func parseCandlestickMsg(conn *websocket.Conn, rawMsg []byte) (*types.CandlestickMsg, error) {
	var arr []interface{}
	err := json.Unmarshal(rawMsg, &arr)
	if err != nil {
		return nil, fmt.Errorf("invalid msg %s", string(rawMsg))
	}
	if len(arr) != 2 {
		return nil, fmt.Errorf("invalid msg %s", string(rawMsg))
	}
	channelId := uint64(arr[0].(float64))
	symbol, ok := idToChannels[channelId]
	if !ok {
		return nil, fmt.Errorf("no channel for %v", channelId)
	}
	base, quote, err := parser.ParseSymbol(exchangeName, symbol)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("rawMsg: %s symbol: %s\n", string(rawMsg), symbol)

	var candles []interface{}
	items, ok := arr[1].([]interface{})
	if !ok {
		heartbeatMsg, ok := arr[1].(string)
		if ok && heartbeatMsg == "hb" {
			conn.WriteJSON(map[string]string{"event": "ping"})
			return nil, nil
		}
		return nil, fmt.Errorf("invalid msg: %v", string(rawMsg))
	}

	// fmt.Printf("first item: %v\n", items[0])
	candles, ok = items[0].([]interface{})
	if !ok {
		candles = items
	}
	if len(candles) != 6 {
		return nil, fmt.Errorf("invalid candles %s", string(rawMsg))
	}

	timestamp := uint64(candles[0].(float64))
	open := candles[1].(float64)
	close := candles[2].(float64)
	high := candles[3].(float64)
	low := candles[4].(float64)
	baseVolume := candles[5].(float64)
	quoteVolume := 0.0

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
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    baseVolume,
		Vwap:      vwap,
	}, nil
}
