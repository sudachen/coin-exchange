package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"time"
)

type Kline struct {
	Timestamp time.Time
	Interval  int32
	TradeNum  int32

	Open   float32
	Close  float32
	High   float32
	Low    float32
	Volume float32
}

type Candlestick struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair

	Kline
}

type Candlesticks struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair

	Klines []Kline
}
