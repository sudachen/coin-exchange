package api

import (
	"github.com/google/logger"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/binance/internal"
	"github.com/sudachen/coin-exchange/exchange/channel"
	"github.com/sudachen/coin-exchange/exchange/ws"
	"strings"
	"sync"
)

type subsid struct {
	channel channel.Channel
	pair    exchange.CoinPair
}

type stream struct {
	endpoint string
	channels []channel.Channel
	pairs    []exchange.CoinPair
	// async mutable part
	ws  *ws.Websocket
	mux *sync.Mutex
}

func (st *stream) Endpoint() string {
	return st.endpoint
}

func (st *stream) OnConnect(wes *ws.Websocket) (bool, error) {
	st.mux.Lock()
	st.ws = wes
	st.mux.Unlock()
	return false, nil
}

func (st *stream) OnDisconnect() {
	st.mux.Lock()
	st.ws = nil
	st.mux.Unlock()
}

func (st *stream) isConnected() bool {
	st.mux.Lock()
	connected := st.ws != nil
	st.mux.Unlock()
	return connected
}

func (st *stream) OnMessage(m []byte) bool {
	switch getChannel(m) {
	case channel.Candlestick:
		if msg, err := internal.CandlestickDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			exchange.Collector.Messages <- msg
			return true
		}
	case channel.Trade:
		if msg, err := internal.TradeDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			exchange.Collector.Messages <- msg
			return true
		}
	case channel.Depth:
		if msg, err := internal.DepthDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			exchange.Collector.Messages <- msg
			return true
		}
	}
	return true
}

func (st *stream) OnFatal(err error) {
	st.mux.Lock()
	st.ws = nil
	st.mux.Unlock()
}

func (st *stream) Close() error {
	st.mux.Lock()
	wes := st.ws
	st.mux.Unlock()
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
	//fmt.Println(s,internal.DepthSuffix)
	if strings.Index(s, "@kline_1m\"") > 0 {
		return channel.Candlestick
	} else if strings.Index(s, "@trade\"") > 0 {
		return channel.Trade
	} else if strings.Index(s, internal.DepthSuffix) > 0 {
		return channel.Depth
	} else {
		return channel.NoChannel
	}
}
