package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"time"
)

type Trade struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair

	TradeId int64
	Price   float32
	Qty     float32

	BuyerOrderId   int64
	SellerOrderId  int64
	TradeOrderTime time.Time
}
