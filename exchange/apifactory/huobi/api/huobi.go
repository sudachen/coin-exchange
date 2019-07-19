package api

import (
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/huobi/internal"
	"github.com/sudachen/coin-exchange/exchange/ws"
	"sync"
	"time"
)

func New() exchange.Api {
	return &api{
		make(map[subsid]bool),
		nil,
		false,
		&sync.Cond{L: &sync.Mutex{}},
	}
}

type api struct {
	subs    map[subsid]bool
	ws      *ws.Websocket
	started bool
	mux     *sync.Cond
}

func (a *api) Lock() {
	a.mux.L.Lock()
}

func (a *api) Unlock() {
	a.mux.L.Unlock()
}

func (a *api) Subscribe(pairs []exchange.CoinPair, channels []exchange.Channel) error {
	a.Lock()
	pairs = a.FilterSupported(pairs)
	for _, c := range channels {
		for _, p := range pairs {
			s := subsid{c, p}
			if _, ok := a.subs[s]; !ok {
				a.subs[s] = false
			}
		}
	}
	if !a.started {
		a.started = true
		ws.Connect(a)
	} else {
		a.mux.Signal()
	}
	a.Unlock()
	return nil
}

func (a *api) IsSupported(pair exchange.CoinPair) bool {
	for _, i := range pair {
		if _, ok := internal.Coins[i]; !ok {
			return false
		}
	}
	return true
}

func (a *api) FilterSupported(pairs []exchange.CoinPair) []exchange.CoinPair {
	var r []exchange.CoinPair
	for _, p := range pairs {
		if a.IsSupported(p) {
			r = append(r, p)
		}
	}
	return r
}

func (a *api) UnsubscribeAll(timeout time.Duration) error {
	a.Lock()
	if a.started && a.ws != nil {
		_ = a.ws.Close()
	}
	a.Unlock()
	return nil
}