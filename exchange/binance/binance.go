package binance

import (
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/binance/rawmesg"
	"strings"
	"time"
)

const baseURL = "wss://stream.binance.com:9443/ws"
const wsTimeout = time.Second * 60
const wsKeepalive = true

func makeSymbol(pair exchange.CoinPair) string {
	var ccs [2]string
	for i, cc := range pair {
		if cc == exchange.USD {
			ccs[i] = "USDT"
		} else {
			ccs[i] = cc.String()
		}
	}
	return strings.ToLower(ccs[0] + ccs[1])
}

func Websocket(pair exchange.CoinPair, channel exchange.Channel) *exchange.Websocket {
	var endpoint string
	switch channel {
	case exchange.Candlestick:
		endpoint = fmt.Sprintf("%s/%s@kline_%s", baseURL, makeSymbol(pair), "1m")
	case exchange.Trade:
		endpoint = fmt.Sprintf("%s/%s@trade", baseURL, makeSymbol(pair))
	default:
		panic("unreachable")
	}
	return exchange.NewWebsocket(
		exchange.StreamId{pair, channel, exchange.Binance},
		endpoint,
		msgConv)
}

func msgConv(sid exchange.StreamId, m []byte) (interface{}, error) {
	switch sid.Channel {
	case exchange.Candlestick:
		if msg, err := rawmesg.CandlelstickDecode(sid, m); err != nil {
			return nil, err
		} else {
			return msg, nil
		}
	case exchange.Trade:
		if msg, err := rawmesg.TradeDecode(sid, m); err != nil {
			return nil, err
		} else {
			return msg, nil
		}
	}
	return nil, nil
}
