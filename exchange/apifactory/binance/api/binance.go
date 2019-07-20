package api

import (
	"fmt"
	"github.com/google/logger"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/binance/internal"
	"github.com/sudachen/coin-exchange/exchange/ws"
	"strings"
	"sync"
	"time"
)

const combinedBaseURL = "wss://stream.binance.com:9443/stream?streams="
const maxEndpointLength = 1000

func New() exchange.Api {
	return &api{
		make(map[subsid]*stream),
		nil,
	}
}

const maxPairsCountInString = 3

func (st *stream) String() string {
	var ss1 []string
	for i, v := range st.pairs {
		if i < maxPairsCountInString {
			ss1 = append(ss1, v.String())
		} else if i == maxPairsCountInString {
			ss1 = append(ss1, "...")
		}
	}
	var ss2 []string
	for _, v := range st.channels {
		ss2 = append(ss2, v.String())
	}
	return "St{Binance|" + strings.Join(ss1, ",") + "|" + strings.Join(ss2, ",") + "}"
}

type api struct {
	subs map[subsid]*stream
	sts  []*stream
}

func (a *api) subscribe(st *stream) {
	for _, channel := range st.channels {
		for _, pair := range st.pairs {
			a.subs[subsid{channel, pair}] = st
		}
	}
	a.sts = append(a.sts, st)
	ws.Connect(st)
}

func (a *api) Subscribe(pairs []exchange.CoinPair, channels []exchange.Channel) error {
	channels = append(channels[:0:0], channels...)
	st := &stream{endpoint: combinedBaseURL, channels: channels, mux: &sync.Mutex{}}
	var sts []*stream

	for _, pair := range a.FilterSupported(pairs) {
		var ep string
		for _, channel := range channels {
			if _, exists := a.subs[subsid{channel, pair}]; !exists {
				switch channel {
				case exchange.Candlestick:
					ep += fmt.Sprintf("%s@kline_%s/", internal.MakeSymbol(pair), "1m")
				case exchange.Trade:
					ep += fmt.Sprintf("%s@trade/", internal.MakeSymbol(pair))
				case exchange.Depth:
					ep += fmt.Sprintf("%s%s/", internal.MakeSymbol(pair), internal.DepthSuffix)
				default:
					panic("unreachable")
				}
			}
		}
		if len(ep)+len(st.endpoint) > maxEndpointLength {
			sts = append(sts, st)
			st = &stream{endpoint: combinedBaseURL, channels: channels, mux: &sync.Mutex{}}
		}
		st.endpoint += ep
		st.pairs = append(st.pairs, pair)
	}

	sts = append(sts, st)

	for _, st := range sts {
		a.subscribe(st)
	}

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

func (a *api) UnsubscribeAll(timeout time.Duration, wg *sync.WaitGroup) {
	wwg := wg
	if wg == nil { wwg = &sync.WaitGroup{} }
	wwg.Add(1)
	go func() {
		hasConneted := true
		startedAt := time.Now()
		a.subs = make(map[subsid]*stream)
		for _, st := range a.sts {
			_ = st.Close()
		}
		for time.Now().Sub(startedAt) < timeout {
			hasConneted = false
			for _, st := range a.sts {
				hasConneted = hasConneted || st.isConnected()
			}
			if !hasConneted {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
		a.sts = a.sts[:0]
		if hasConneted {
			logger.Errorf("Binance API still has connected streams")
		}
		wwg.Done()
	}()
	if wg == nil { wwg.Wait() }
}

