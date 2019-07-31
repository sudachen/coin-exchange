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

type Limit int

const (
	LimitMin Limit = iota
	LimitLow
	LimitNorm
	LimitHigh
	LimitMax
)

type QueryApi interface {
	QueryDepth(Limit) (*Orders, error)
	QueryTrades(Limit) (*Trades, error)
	QueryAggTrades(Limit) (*Trades, error)
	QueryCandlesticks(interval int, count int) (*Candlesticks, error)
}
