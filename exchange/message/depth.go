package message

import (
	"github.com/sudachen/coin-exchange/exchange"
)

type DepthValue struct {
	Price float32
	Qty   float32
}

type Depth struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair

	FirstUpdateId int64
	LastUpdateId  int64
	Bids          []DepthValue
	Asks          []DepthValue
}
