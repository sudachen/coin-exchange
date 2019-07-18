package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"time"
)

type DepthValue struct {
	Price float32
	Qty   float32
}

type Depth struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair
	Bids   []DepthValue
	Asks   []DepthValue

	Timestamp time.Time
}
