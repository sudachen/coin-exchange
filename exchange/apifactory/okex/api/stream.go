package api

import (
	"encoding/json"
	"github.com/google/logger"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/okex/internal"
	"github.com/sudachen/coin-exchange/exchange/channel"
	"github.com/sudachen/coin-exchange/exchange/ws"
	"strings"
)

type subsid struct {
	channel channel.Channel
	pair    exchange.CoinPair
}

func (a *api) Endpoint() string {
	return "wss://real.okex.com:10442/ws/v3?compress=true"
}

type query struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}

var channels = []channel.Channel{channel.Candlestick, channel.Trade, channel.Depth}

func (a *api) subscribeAll() {
	//logger.Info("Okex ws subscriber started")
	for a.isConnected() {
		for _, c := range channels {
			var args []string
			var pfx string
			switch c {
			case channel.Candlestick:
				pfx = "spot/candle60s:"
			case channel.Trade:
				pfx = "spot/trade:"
			case channel.Depth:
				pfx = "spot/depth5:"
			}
			a.Lock()
			for k, ready := range a.subs {
				if !ready && k.channel == c {
					args = append(args, pfx+internal.MakeSymbol(k.pair))
					a.subs[k] = true
				}
			}
			wes := a.ws
			a.Unlock()
			if len(args) > 0 {
				q := &query{"subscribe", args}
				bs, _ := json.Marshal(q)
				_ = wes.Send(bs)
			}
		}
		a.Lock()
		a.mux.Wait()
		a.Unlock()
	}
	//logger.Info("Okex ws subscriber finished")
}

func (a *api) OnConnect(wes *ws.Websocket) (bool, error) {
	a.Lock()
	a.ws = wes
	for k, _ := range a.subs {
		a.subs[k] = false
	}
	go a.subscribeAll()
	a.ws.KeepAlive(func(wes *ws.Websocket) error {
		err := wes.Send([]byte("ping"))
		return err
	})
	a.Unlock()
	return false, nil
}

func (a *api) OnDisconnect() {
	a.Lock()
	a.ws = nil
	a.mux.Signal()
	a.Unlock()
}

func (a *api) isConnected() bool {
	a.Lock()
	connected := a.ws != nil
	a.Unlock()
	return connected
}

func (a *api) OnMessage(m []byte) bool {
	switch getChannel(m) {
	case channel.Candlestick:
		if msg, err := internal.CandlestickDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			for _, m := range msg {
				exchange.Collector.Messages <- m
			}
			return true
		}
	case channel.Trade:
		if msg, err := internal.TradeDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			for _, m := range msg {
				exchange.Collector.Messages <- m
			}
			return true
		}
	case channel.Depth:
		if msg, err := internal.DepthDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			for _, m := range msg {
				exchange.Collector.Messages <- m
			}
			return true
		}
	}
	return true
}

func (a *api) OnFatal(err error) {
	a.Lock()
	a.ws = nil
	a.mux.Signal()
	a.Unlock()
}

func (a *api) Close() error {
	a.Lock()
	wes := a.ws
	a.Unlock()
	if wes != nil {
		return wes.Close()
	} else {
		return nil
	}
}

func getChannel(m []byte) channel.Channel {
	if len(m) > 80 {
		m = m[:80]
	}
	s := string(m)
	if strings.Index(s, "\"spot/candle60s\"") > 0 {
		return channel.Candlestick
	} else if strings.Index(s, "\"spot/trade\"") > 0 {
		return channel.Trade
	} else if strings.Index(s, "\"spot/depth5\"") > 0 {
		return channel.Depth
	} else {
		return channel.NoChannel
	}
}

const maxPairsCountInString = 3

func (a *api) String() string {
	a.Lock()
	cls := make(map[channel.Channel][]string)
	for k, _ := range a.subs {
		ss1, ok := cls[k.channel]
		if !ok {
			ss1 = make([]string, 0, 3)
		}
		if len(ss1) < maxPairsCountInString {
			ss1 = append(ss1, k.pair.String())
		} else if len(ss1) == maxPairsCountInString {
			ss1 = append(ss1, "...")
		}
		cls[k.channel] = ss1
	}
	var ss2 []string
	for k, v := range cls {
		ss2 = append(ss2, k.String()+":"+strings.Join(v, ","))
	}
	a.Unlock()
	return "St{Okex|" + strings.Join(ss2, ";") + "}"
}
