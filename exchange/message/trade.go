package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"time"
)

type Trade struct {
	Origin    exchange.Exchange
	Pair      exchange.CoinPair
	Price     float32
	Qty       float32
	Sell      bool
	Timestamp time.Time
}
