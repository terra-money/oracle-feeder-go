package huobi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	websocketUrl string = "wss://api.huobi.pro/ws"
	exchangeName string = "huobi"
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

	for _, symbol := range symbols {
		command := generateCommand(symbol)
		if err := conn.WriteJSON(&command); err != nil {
			return nil, err
		}

		_, compressed, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		gzreader, _ := gzip.NewReader(bytes.NewReader(compressed))
		decompressed, _ := io.ReadAll(gzreader)

		var resp map[string]any
		err = json.Unmarshal(decompressed, &resp)
		if err != nil {
			return nil, err
		}

		if status, ok := resp["status"].(string); ok && status != "ok" {
			bytes, _ := json.Marshal(resp)
			return nil, fmt.Errorf("%s", string(bytes))
		}
	}

	return conn, nil
}

// Candlestick websocket message.
//
// Message format: https://huobiapi.github.io/docs/spot/v1/en/#market-candlestick
type RawCandlestickMsg struct {
	Channel   string `json:"ch"`
	Timestamp uint64 `json:"ts"`
	Tick      struct {
		Id          uint64  `json:"id"`
		Open        float64 `json:"open"`
		High        float64 `json:"high"`
		Low         float64 `json:"low"`
		Close       float64 `json:"close"`
		Volume      float64 `json:"amount"`
		QuoteVolume float64 `json:"vol"`
	} `json:"tick"`
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://huobiapi.github.io/docs/spot/v1/en/#market-candlestick
//
// For example:
// {"sub":"market.btcusdt.kline.1min","id":"terra-price-server"}
func generateCommand(symbol string) map[string]interface{} {
	sub := fmt.Sprintf("market.%s.kline.1min", strings.ToLower(symbol))
	return map[string]interface{}{
		"sub": sub,
		"id":  "terra-price-server",
	}
}

func (wc *WebsocketClient) ParseCandlestickMsg(rawMsg []byte) (*types.CandlestickMsg, error) {
	var msg RawCandlestickMsg
	json.Unmarshal(rawMsg, &msg)

	arr := strings.Split(msg.Channel, ".")
	if len(arr) != 4 {
		return nil, fmt.Errorf("not a candlestick %s", string(rawMsg))
	}
	symbol := arr[1]
	base, quote, err := parser.ParseSymbol(exchangeName, symbol)
	if err != nil {
		return nil, err
	}
	vwap := 0.0
	if msg.Tick.Volume == 0.0 || msg.Tick.QuoteVolume == 0.0 {
		vwap = (msg.Tick.Open + msg.Tick.Close) / 2.0
	} else {
		vwap = msg.Tick.QuoteVolume / msg.Tick.Volume
	}
	return &types.CandlestickMsg{
		Exchange:  exchangeName,
		Symbol:    symbol,
		Base:      base,
		Quote:     quote,
		Timestamp: msg.Timestamp,
		Open:      msg.Tick.Open,
		High:      msg.Tick.High,
		Low:       msg.Tick.Low,
		Close:     msg.Tick.Close,
		Volume:    msg.Tick.Volume,
		Vwap:      vwap,
	}, nil
}
