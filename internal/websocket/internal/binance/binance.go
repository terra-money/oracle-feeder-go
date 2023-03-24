package binance

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	websocketUrl string = "wss://stream.binance.com:9443/stream"
	exchangeName string = "binance"
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

	command := generateCommand(symbols)
	if err := conn.WriteJSON(&command); err != nil {
		return nil, err
	}

	resp := make(map[string]interface{})
	if err := conn.ReadJSON(&resp); err != nil {
		return nil, err
	}
	if _, ok := resp["error"]; ok {
		bytes, _ := json.Marshal(resp)
		return nil, fmt.Errorf("%s", string(bytes))
	}

	return conn, nil
}

func (wc *WebsocketClient) HandleMsg(msg []byte, conn *websocket.Conn) (*types.CandlestickMsg, error) {
	return parseCandlestickMsg(msg)
}

// Candlestick websocket message.
//
// Message format: https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-streams
type RawCandlestickMsg struct {
	Stream string `json:"stream"`
	Data   struct {
		EventTime uint64         `json:"E"`
		EventType string         `json:"e"`
		Symbol    string         `json:"s"`
		Kline     map[string]any `json:"k"`
	} `json:"data"`
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-streams
//
// For example:
// {"id":9527,"method":"SUBSCRIBE","params":["btcusdt@kline_1m","ethusdt@kline_1m"]}
func generateCommand(symbols []string) map[string]interface{} {
	var params []string
	for _, symbol := range symbols {
		params = append(params, fmt.Sprintf("%s@kline_1m", strings.ToLower(symbol)))
	}
	return map[string]interface{}{
		"id":     9527,
		"method": "SUBSCRIBE",
		"params": params,
	}
}

func parseCandlestickMsg(rawMsg []byte) (*types.CandlestickMsg, error) {
	var msg RawCandlestickMsg
	json.Unmarshal(rawMsg, &msg)
	symbol := msg.Data.Kline["s"].(string)
	base, quote, err := parser.ParseSymbol(exchangeName, symbol)
	if err != nil {
		return nil, err
	}

	baseVolume, err := strconv.ParseFloat(msg.Data.Kline["v"].(string), 64)
	if err != nil {
		return nil, err
	}
	quoteVolume, err := strconv.ParseFloat(msg.Data.Kline["q"].(string), 64)
	if err != nil {
		return nil, err
	}
	open, err := strconv.ParseFloat(msg.Data.Kline["o"].(string), 64)
	if err != nil {
		return nil, err
	}
	high, err := strconv.ParseFloat(msg.Data.Kline["h"].(string), 64)
	if err != nil {
		return nil, err
	}
	low, err := strconv.ParseFloat(msg.Data.Kline["l"].(string), 64)
	if err != nil {
		return nil, err
	}
	close, err := strconv.ParseFloat(msg.Data.Kline["c"].(string), 64)
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
		Timestamp: uint64(msg.Data.Kline["T"].(float64)),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    baseVolume,
		Vwap:      vwap,
	}, nil
}
