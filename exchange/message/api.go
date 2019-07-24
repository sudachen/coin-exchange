package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/channel"
	"sync"
	"time"
)

type Api interface {
	Subscribe([]exchange.CoinPair, ...channel.Channel) error
	IsSupported(exchange.CoinPair) bool
	FilterSupported([]exchange.CoinPair) []exchange.CoinPair
	UnsubscribeAll(time.Duration, *sync.WaitGroup /*can be nil*/)
	Queries(exchange.CoinPair) (QueryApi, error)
}

type QueryApi interface {
	QueryDepth(count int32) (*Orders, error)
	QueryTrades(count int32) (*Trades, error)
	QueryCandlesticks(interval int32, count int32) (*Candlesticks, error)
}

