package okx

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	websocketUrl string = "wss://ws.okx.com:8443/ws/v5/public"
	exchangeName string = "okx"
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

	// send ping per 15 seconds
	// see https://www.okx.com/docs-v5/en/#websocket-api-connect
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			err = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Printf("%v", err)
			}
		}
	}()

	command, err := generateCommand(symbols)
	if err := conn.WriteJSON(&command); err != nil {
		return nil, err
	}

	return conn, nil
}

func (wc *WebsocketClient) HandleMsg(msg []byte, conn *websocket.Conn) (*types.CandlestickMsg, error) {
	if string(msg) == "pong" {
		return nil, nil
	}
	resp := make(map[string]interface{})
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, err
	}

	if event, ok := resp["event"].(string); ok {
		if event == "error" {
			return nil, fmt.Errorf("%s", string(msg))
		} else {
			return nil, nil
		}
	}

	_, argExist := resp["arg"].(map[string]interface{})
	_, dataExist := resp["data"].([]interface{})
	if argExist && dataExist {
		return parseCandlestickMsg(msg)
	} else {
		return nil, nil
	}
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://www.okx.com/docs-v5/en/#websocket-api-public-channel-candlesticks-channel
//
// For example:
// {"op":"subscribe","args":[{"channel":"candle1m","instId":"BTC-USDT"},{"channel":"candle1m","instId":"ETH-USDT"}]}
func generateCommand(symbols []string) (map[string]interface{}, error) {
	var args []map[string]string
	for _, symbol := range symbols {
		arg := map[string]string{
			"channel": "candle1m",
			"instId":  symbol,
		}
		args = append(args, arg)
	}
	return map[string]interface{}{
		"op":   "subscribe",
		"args": args,
	}, nil
}

// https://docs.kraken.com/websockets/#message-ohlc
func parseCandlestickMsg(rawMsg []byte) (*types.CandlestickMsg, error) {
	resp := make(map[string]interface{})
	err := json.Unmarshal(rawMsg, &resp)
	if err != nil {
		return nil, fmt.Errorf("invalid candlestick %s", string(rawMsg))
	}
	data := resp["data"].([]interface{})
	if len(data) != 1 {
		return nil, fmt.Errorf("no data %s", string(rawMsg))
	}

	arg := resp["arg"].(map[string]interface{})
	symbol := arg["instId"].(string)

	candles := data[0].([]interface{})

	base, quote, err := parser.ParseSymbol(exchangeName, symbol)
	if err != nil {
		return nil, err
	}

	if len(candles) != 9 {
		return nil, fmt.Errorf("invalid candles %s", string(rawMsg))
	}

	timestamp, err := strconv.ParseFloat(candles[0].(string), 64)
	if err != nil {
		return nil, err
	}
	open, err := strconv.ParseFloat(candles[1].(string), 64)
	if err != nil {
		return nil, err
	}
	high, err := strconv.ParseFloat(candles[2].(string), 64)
	if err != nil {
		return nil, err
	}
	low, err := strconv.ParseFloat(candles[3].(string), 64)
	if err != nil {
		return nil, err
	}
	close, err := strconv.ParseFloat(candles[4].(string), 64)
	if err != nil {
		return nil, err
	}
	baseVolume, err := strconv.ParseFloat(candles[5].(string), 64)
	if err != nil {
		return nil, err
	}
	quoteVolume, err := strconv.ParseFloat(candles[6].(string), 64)
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
		Timestamp: uint64(timestamp),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    baseVolume,
		Vwap:      vwap,
	}, nil
}
