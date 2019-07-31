package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/logger"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/huobi/internal"
	"github.com/sudachen/coin-exchange/exchange/channel"
	"github.com/sudachen/coin-exchange/exchange/ws"
	"strings"
	"time"
)

type subsid struct {
	channel channel.Channel
	pair    exchange.CoinPair
}

func (a *api) Endpoint() string {
	return "wss://api.huobi.pro/ws"
}

type query struct {
	Sub string `json:"sub"`
	Id  string `json:"id"`
}

var channels = []channel.Channel{channel.Candlestick, channel.Trade, channel.Depth}
var idCounter = 0

func (a *api) subscribeAll() {
	for a.isConnected() {
		for _, c := range channels {
			var cs []string
			var sfx string
			switch c {
			case channel.Candlestick:
				sfx = ".kline.1min"
			case channel.Trade:
				sfx = ".trade.detail"
			case channel.Depth:
				sfx = ".depth.step1"
			}
			a.Lock()
			for k, ready := range a.subs {
				if !ready && k.channel == c {
					cs = append(cs, "market."+internal.MakeSymbol(k.pair)+sfx)
					a.subs[k] = true
				}
			}
			wes := a.ws
			a.Unlock()
			for _, c := range cs {
				idCounter += 1
				q := &query{c, fmt.Sprintf("%d", idCounter)}
				bs, _ := json.Marshal(q)
				_ = wes.Send(bs)
			}
		}
		a.Lock()
		a.mux.Wait()
		a.Unlock()
	}
}

func (a *api) OnConnect(wes *ws.Websocket) (bool, error) {
	a.Lock()
	a.ws = wes
	wes.Compression = ws.Gzipped
	for k, _ := range a.subs {
		a.subs[k] = false
	}
	go a.subscribeAll()
	a.ws.KeepAlive(func(wes *ws.Websocket) error {
		err := wes.Send([]byte(fmt.Sprintf("{\"ping\":%v}", time.Now().UnixNano()/int64(time.Millisecond))))
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

	e := make(map[string]interface{})
	if err := json.Unmarshal(m, &e); err == nil {
		//logger.Infof("%#v",e)
		if t, ok := e["ping"]; ok {
			pong := fmt.Sprintf("{\"pong\":%v}", t)
			//logger.Info(pong)
			_ = a.ws.Send([]byte(pong))
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
	if strings.Index(s, "\"status\":\"") > 0 {
		return channel.NoChannel
	}
	if strings.Index(s, ".kline.1min\"") > 0 {
		return channel.Candlestick
	} else if strings.Index(s, ".trade.detail\"") > 0 {
		return channel.Trade
	} else if strings.Index(s, ".depth.") > 0 {
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
	return "St{Huobi|" + strings.Join(ss2, ";") + "}"
}
