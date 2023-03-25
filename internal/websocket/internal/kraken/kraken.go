package kraken

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	websocketUrl string = "wss://ws.kraken.com"
	exchangeName string = "kraken"
)

type WebsocketClient struct{}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{}
}

func (wc *WebsocketClient) ConnectAndSubscribe(symbols []string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl, nil)
	if err != nil {
		return nil, err
	}

	commands := generateCommands(symbols)
	for _, command := range commands {
		if err := conn.WriteJSON(&command); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func (wc *WebsocketClient) HandleMsg(msg []byte, conn *websocket.Conn) (*types.CandlestickMsg, error) {
	if strings.HasPrefix(string(msg), "[") {
		return parseCandlestickMsg(msg)
	}
	if strings.HasPrefix(string(msg), "{") {
		resp := make(map[string]interface{})
		if err := json.Unmarshal(msg, &resp); err != nil {
			return nil, err
		}
		event, ok := resp["event"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid msg: %s", string(msg))
		}
		if event == "heartbeat" {
			return nil, nil
		}
		if event == "systemStatus" || event == "subscriptionStatus" {
			log.Printf("Received msg: %v\n", string(msg))
			return nil, nil
		}
	}
	log.Printf("Unrecognized msg: %v\n", string(msg))
	return nil, nil
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://docs.kraken.com/websockets/#message-subscribe
//
// Kucoin allows to subscribe 100 symbols per time, otherwise you'll get an error
// "exceed max subscription count limitation of 100 per time"
//
// For example:
// {"event": "subscribe","pair": ["XBT/EUR"],"subscription": {"interval": 5,"name": "ohlc"}}
func generateCommand(symbols []string) (map[string]interface{}, error) {
	if len(symbols) > 100 {
		return nil, fmt.Errorf("exceeds 100 symbols")
	}
	var topics []string
	for _, symbol := range symbols {
		topics = append(topics, symbol)
	}

	subscription := map[string]interface{}{
		"interval": 1,
		"name":     "ohlc",
	}
	return map[string]interface{}{
		"event":        "subscribe",
		"pair":         topics,
		"subscription": subscription,
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

// https://docs.kraken.com/websockets/#message-ohlc
func parseCandlestickMsg(rawMsg []byte) (*types.CandlestickMsg, error) {
	var arr []interface{}
	err := json.Unmarshal(rawMsg, &arr)
	if err != nil {
		return nil, fmt.Errorf("Invalid msg %s", string(rawMsg))
	}
	if len(arr) != 4 {
		return nil, fmt.Errorf("Invalid msg %s", string(rawMsg))
	}
	// channelId := arr[0].(float64)
	candles := arr[1].([]interface{})
	name := arr[2].(string)
	symbol := arr[3].(string)

	items := strings.Split(name, "-")
	if len(items) > 0 && items[0] != "ohlc" {
		return nil, fmt.Errorf("Not ohlc %s", string(rawMsg))
	}

	base, quote, err := parser.ParseSymbol(exchangeName, symbol)
	if err != nil {
		return nil, err
	}

	if len(candles) != 9 {
		return nil, fmt.Errorf("Invalid candles %s", string(rawMsg))
	}

	timestampSeconds, err := strconv.ParseFloat(candles[0].(string), 64)
	if err != nil {
		return nil, err
	}
	timestamp := uint64(timestampSeconds * 1e3)
	open, err := strconv.ParseFloat(candles[2].(string), 64)
	if err != nil {
		return nil, err
	}
	high, err := strconv.ParseFloat(candles[3].(string), 64)
	if err != nil {
		return nil, err
	}
	low, err := strconv.ParseFloat(candles[4].(string), 64)
	if err != nil {
		return nil, err
	}
	close, err := strconv.ParseFloat(candles[5].(string), 64)
	if err != nil {
		return nil, err
	}
	vwap, err := strconv.ParseFloat(candles[6].(string), 64)
	if err != nil {
		return nil, err
	}
	baseVolume, err := strconv.ParseFloat(candles[7].(string), 64)
	if err != nil {
		return nil, err
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
