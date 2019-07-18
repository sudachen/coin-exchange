package api

import (
	"github.com/google/logger"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/binance/internal"
	"github.com/sudachen/coin-exchange/exchange/ws"
	"strings"
	"sync"
)

type subsid struct {
	channel exchange.Channel
	pair    exchange.CoinPair
}

type stream struct {
	endpoint string
	channels []exchange.Channel
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
	case exchange.Candlestick:
		if msg, err := internal.CandlestickDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			exchange.Collector.Messages <- msg
			return true
		}
	case exchange.Trade:
		if msg, err := internal.TradeDecode(m); err != nil {
			logger.Error(err.Error())
			return false
		} else {
			exchange.Collector.Messages <- msg
			return true
		}
	case exchange.Depth:
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

func getChannel(m []byte) exchange.Channel {
	s := string(m)
	if strings.Index(s, "@kline_1m\"") > 0 {
		return exchange.Candlestick
	} else if strings.Index(s, "@trade\"") > 0 {
		return exchange.Trade
	} else if strings.Index(s, "@depth5\"") > 0 {
		return exchange.Depth
	} else {
		return exchange.NoChannel
	}
}
