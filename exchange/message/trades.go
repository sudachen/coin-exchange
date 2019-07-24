package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"time"
)

type TradeValue struct {
	Price     float32
	Qty       float32
	Sell      bool
	Timestamp time.Time
}

type Trade struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair
	Value  TradeValue
}

type Trades struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair
	Values []TradeValue
}
