package coinbase

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/parser"
	"github.com/terra-money/oracle-feeder-go/internal/types"
)

const (
	websocketUrl string = "wss://ws-feed.exchange.coinbase.com"
	exchangeName string = "coinbase"
	INTERVAL     uint64 = 60 * 1000
)

var symbolToBarTime = map[string]uint64{}
var symbolToCandle = map[string]*types.CandlestickMsg{}

type WebsocketClient struct{}

func NewWebsocketClient() *WebsocketClient {
	return &WebsocketClient{}
}

// Trade websocket message.
//
// Message format: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#match
type RawTradeMsg struct {
	Type         uint64 `json:"type"`
	TradeId      string `json:"trade_id"`
	Sequence     string `json:"sequence"`
	MakerOrderId string `json:"maker_order_id"`
	TakerOrderId string `json:"taker_order_id"`
	Time         string `json:"time"`
	ProductId    string `json:"product_id"`
	Size         string `json:"size"`
	Price        string `json:"price"`
	Side         string `json:"side"`
}

type TradeMsg struct {
	Exchange  string  // Exchange name
	Symbol    string  // Exchange-specific trading symbol
	Base      string  // Base coin
	Quote     string  // Quote coin
	Timestamp uint64  // Trade timestamp
	Price     float64 // Trade price
	Volume    float64 // Base Volume
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

	return conn, nil
}

func (wc *WebsocketClient) HandleMsg(msg []byte, conn *websocket.Conn) (*types.CandlestickMsg, error) {
	resp := make(map[string]interface{})
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, err
	}

	typ, ok := resp["type"].(string)
	if !ok {
		return nil, fmt.Errorf("no type: %s", string(msg))
	}
	if typ == "error" {
		return nil, fmt.Errorf("error: %s", string(msg))
	}
	if typ == "subscriptions" {
		return nil, nil
	}

	if typ == "match" || typ == "last_match" {
		tradeMsg, err := parseTradeMsg(msg)
		if err != nil {
			return nil, err
		}
		return generateCandleStickMsg(tradeMsg), nil
	}
	return nil, fmt.Errorf("invalid msg: %s", string(msg))
}

// generateCommand generates the trade subscription command from specified symbols.
//
// API doc: https://docs.cloud.coinbase.com/exchange/docs/websocket-channels#matches-channel
//
// For example:
// {"type":"subscribe","channels": [{"name":"matches","product_ids":["BTC-USD","ETH-USD"]}]}
func generateCommand(symbols []string) map[string]interface{} {
	channel := map[string]interface{}{
		"name":        "matches",
		"product_ids": symbols,
	}
	var channels []interface{}
	channels = append(channels, channel)
	return map[string]interface{}{
		"type":     "subscribe",
		"channels": channels,
	}
}

func parseTradeMsg(rawMsg []byte) (*TradeMsg, error) {
	var msg RawTradeMsg
	json.Unmarshal(rawMsg, &msg)
	base, quote, err := parser.ParseSymbol(exchangeName, msg.ProductId)
	if err != nil {
		return nil, err
	}
	price, err := strconv.ParseFloat(msg.Price, 64)
	if err != nil {
		return nil, err
	}
	volume, err := strconv.ParseFloat(msg.Size, 64)
	if err != nil {
		return nil, err
	}
	timeObj, err := time.Parse(time.RFC3339, msg.Time)
	if err != nil {
		return nil, err
	}
	timestamp := uint64(timeObj.UnixMilli())
	tradeMsg := &TradeMsg{
		Exchange:  exchangeName,
		Symbol:    msg.ProductId,
		Base:      base,
		Quote:     quote,
		Price:     price,
		Volume:    volume,
		Timestamp: timestamp,
	}
	return tradeMsg, nil
}

func generateCandleStickMsg(tradeMsg *TradeMsg) *types.CandlestickMsg {
	var candle *types.CandlestickMsg
	lastBarTime := symbolToBarTime[tradeMsg.Symbol]
	nextBarTime := lastBarTime + INTERVAL
	if tradeMsg.Timestamp >= nextBarTime {
		symbolToBarTime[tradeMsg.Symbol] = nextBarTime
		candle = symbolToCandle[tradeMsg.Symbol]
		symbolToCandle[tradeMsg.Symbol] = nil
	}
	addTrade(tradeMsg)
	return candle
}

func addTrade(trade *TradeMsg) error {
	symbol := trade.Symbol
	base := trade.Base
	quote := trade.Quote
	candle := symbolToCandle[symbol]
	timestamp := trade.Timestamp
	baseVolume := trade.Volume
	quoteVolume := trade.Price * trade.Volume
	if candle == nil {
		open := trade.Price
		close := trade.Price
		high := trade.Price
		low := trade.Price
		vwap := 0.0
		if baseVolume == 0.0 || quoteVolume == 0.0 {
			vwap = (open + close) / 2.0
		} else {
			vwap = quoteVolume / baseVolume
		}
		candle = &types.CandlestickMsg{
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
		}
		symbolToCandle[symbol] = candle
		return nil
	}
	if candle.High < trade.Price {
		candle.High = trade.Price
	}
	if candle.Low > trade.Price {
		candle.Low = trade.Price
	}
	candle.Close = trade.Price
	prevBaseVolume := candle.Volume
	candle.Volume += trade.Volume
	candle.Vwap = (prevBaseVolume*candle.Vwap + quoteVolume) / candle.Volume
	return nil
}
