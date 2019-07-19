package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"time"
)

type Candlestick struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair

	Timestamp time.Time
	Interval  int32
	TradeNum  int32

	Open   float32
	Close  float32
	High   float32
	Low    float32
	Volume float32
}
