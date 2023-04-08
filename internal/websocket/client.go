package websocket

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/terra-money/oracle-feeder-go/internal/types"
	"github.com/terra-money/oracle-feeder-go/internal/websocket/internal/binance"
	"github.com/terra-money/oracle-feeder-go/internal/websocket/internal/bitfinex"
	"github.com/terra-money/oracle-feeder-go/internal/websocket/internal/coinbase"
	"github.com/terra-money/oracle-feeder-go/internal/websocket/internal/huobi"
	"github.com/terra-money/oracle-feeder-go/internal/websocket/internal/kraken"
	"github.com/terra-money/oracle-feeder-go/internal/websocket/internal/kucoin"
	"github.com/terra-money/oracle-feeder-go/internal/websocket/internal/okx"
)

type websocketClient interface {
	ConnectAndSubscribe(symbols []string) (*websocket.Conn, error)
	// HandleMsg handles websocket messages and returns a CandlestickMsg if possible
	HandleMsg(msg []byte, conn *websocket.Conn) (*types.CandlestickMsg, error)
}

// SubscribeCandlestick subscribes to the candlestick channel.
func SubscribeCandlestick(exchange string, symbols []string, stopCh <-chan struct{}) (<-chan *types.CandlestickMsg, error) {
	var client websocketClient
	switch strings.ToLower(exchange) {
	case "binance":
		client = binance.NewWebsocketClient()
	case "coinbase":
		client = coinbase.NewWebsocketClient()
	case "huobi":
		client = huobi.NewWebsocketClient()
	case "kucoin":
		client = kucoin.NewWebsocketClient()
	case "bitfinex":
		client = bitfinex.NewWebsocketClient()
	case "kraken":
		client = kraken.NewWebsocketClient()
	case "okx":
		client = okx.NewWebsocketClient()
	default:
		return nil, fmt.Errorf("unknown websocket exchange %s", exchange)
	}

	conn, err := client.ConnectAndSubscribe(symbols)
	if err != nil {
		return nil, err
	}

	outCh := make(chan *types.CandlestickMsg)

	go func() {
		var rawMsg []byte
		for {
			select {
			case <-stopCh:
				conn.Close()
				close(outCh)
				return
			default:
				if conn != nil {
					_, rawMsg, err = conn.ReadMessage()
				}

				if err == nil {
					candlestick, err := client.HandleMsg(rawMsg, conn)
					if err != nil {
						log.Printf("%v", err)
					}
					if candlestick != nil {
						outCh <- candlestick
					}
				} else {
					if websocket.IsCloseError(err) {
						// reconnect automatically
						conn, err = client.ConnectAndSubscribe(symbols)
						if err != nil {
							log.Printf("%v", err)
							time.Sleep(3 * time.Second)
						}
					}
				}
			}
		}
	}()

	return outCh, nil
}
