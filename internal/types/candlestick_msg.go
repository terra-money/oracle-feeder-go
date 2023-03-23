package types

// CandlestickMsg represents an OHLCV message.
type CandlestickMsg struct {
	Exchange  string  // Exchange name
	Symbol    string  // Exchange-specific trading symbol
	Base      string  // Base coin
	Quote     string  // Quote Coin
	Timestamp uint64  // Kline close time
	Open      float64 // open price
	High      float64 // high price
	Low       float64 // low price
	Close     float64 // close price
	Volume    float64 // base volume
	Vwap      float64 // volume weighted average price
}
