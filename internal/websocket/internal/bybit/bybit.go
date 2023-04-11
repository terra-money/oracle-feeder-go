package bybit

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	// API doc: https://bybit-exchange.github.io/docs/v5/ws/connect
	websocketUrl string = "wss://stream.bybit.com/v5/public/spot"
	exchangeName string = "bybit"
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

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			pingCommand := generatePingCommand()
			if err := conn.WriteJSON(&pingCommand); err != nil {
				log.Printf("%v", err)
			}
		}
	}()

	return conn, nil
}

func (wc *WebsocketClient) HandleMsg(msg []byte, conn *websocket.Conn) (*types.CandlestickMsg, error) {
	resp := make(map[string]interface{})
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, err
	}

	if op, ok := resp["op"].(string); ok {
		if op == "ping" || op == "subscribe" {
			success := resp["success"].(bool)
			if !success {
				return nil, fmt.Errorf("msg error: %v", string(msg))
			} else {
				return nil, nil
			}
		} else {
			log.Printf("receive msg: %s\n", string(msg))
			return nil, nil
		}
	}

	if _, topicExists := resp["topic"].(string); !topicExists {
		return nil, fmt.Errorf("no topic in %s", string(msg))
	}
	if _, dataExists := resp["data"].([]interface{}); !dataExists {
		return nil, fmt.Errorf("no data in %s", string(msg))
	}

	return parseCandlestickMsg(msg)
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://bybit-exchange.github.io/docs/v5/websocket/public/kline
//
// For example:
// {"req_id": "terra-price-server","op":"subscribe","args":["kline.1.BTCUSDT", "kline.1.ETHUSDT"]}
func generateCommand(symbols []string) map[string]interface{} {
	var args []string
	for _, symbol := range symbols {
		arg := fmt.Sprintf("kline.1.%s", symbol)
		args = append(args, arg)
	}
	return map[string]interface{}{
		"req_id": "terra-price-server",
		"op":     "subscribe",
		"args":   args,
	}
}

// Spot can input up to 10 args for each subscription request sent to one connection,
// see https://bybit-exchange.github.io/docs/v5/ws/connect#public-channel---args-limits
func generateCommands(symbols []string) []map[string]interface{} {
	var commands []map[string]interface{}
	groupSize := 10
	n := len(symbols)
	for i := 0; i < n; i += groupSize {
		j := i + groupSize
		if j > n {
			j = n
		}
		group := symbols[i:j]
		commands = append(commands, generateCommand(group))
	}
	return commands
}

// generateCommand generates the candlestick subscription command from specified symbols.
//
// API doc: https://bybit-exchange.github.io/docs/v5/ws/connect#how-to-send-the-heartbeat-packet
//
// For example:
// {"req_id": "100001", "op": "ping"}
func generatePingCommand() map[string]string {
	return map[string]string{
		"req_id": "terra-price-server",
		"op":     "ping",
	}
}

// https://bybit-exchange.github.io/docs/v5/websocket/public/kline
func parseCandlestickMsg(rawMsg []byte) (*types.CandlestickMsg, error) {
	resp := make(map[string]interface{})
	err := json.Unmarshal(rawMsg, &resp)
	if err != nil {
		return nil, fmt.Errorf("invalid candlestick %s", string(rawMsg))
	}

	data := resp["data"].([]interface{})
	if len(data) != 1 {
		return nil, fmt.Errorf("invalid data %s", string(rawMsg))
	}

	topic := resp["topic"].(string)
	items := strings.Split(topic, ".")
	if len(items) != 3 {
		return nil, fmt.Errorf("invalid topic %s", topic)
	}
	symbol := items[2]

	candles := data[0].(map[string]interface{})

	base, quote, err := parser.ParseSymbol(exchangeName, symbol)
	if err != nil {
		return nil, err
	}

	confirm := candles["confirm"].(bool)
	if !confirm {
		// this candle is not closed yet
		return nil, nil
	}

	timestamp := uint64(candles["end"].(float64))
	open, err := strconv.ParseFloat(candles["open"].(string), 64)
	if err != nil {
		return nil, err
	}
	high, err := strconv.ParseFloat(candles["high"].(string), 64)
	if err != nil {
		return nil, err
	}
	low, err := strconv.ParseFloat(candles["low"].(string), 64)
	if err != nil {
		return nil, err
	}
	close, err := strconv.ParseFloat(candles["close"].(string), 64)
	if err != nil {
		return nil, err
	}
	baseVolume, err := strconv.ParseFloat(candles["volume"].(string), 64)
	if err != nil {
		return nil, err
	}
	quoteVolume, err := strconv.ParseFloat(candles["turnover"].(string), 64)
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
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    baseVolume,
		Vwap:      vwap,
	}, nil
}
