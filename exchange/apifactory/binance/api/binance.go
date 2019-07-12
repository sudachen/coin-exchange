package api

import (
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/binance/internal"
	"time"
)

const baseURL = "wss://stream.binance.com:9443/ws"
const combinedBaseURL = "wss://stream.binance.com:9443/stream?streams="
const wsTimeout = time.Second * 60
const wsKeepalive = true

type Api struct{}
type Conv struct{ combined bool }

func (_ *Api) Subscribe(pair exchange.CoinPair, channel exchange.Channel) error {
	var endpoint string
	switch channel {
	case exchange.Candlestick:
		endpoint = fmt.Sprintf("%s/%s@kline_%s", baseURL, internal.MakeSymbol(pair), "1m")
	case exchange.Trade:
		endpoint = fmt.Sprintf("%s/%s@trade", baseURL, internal.MakeSymbol(pair))
	default:
		panic("unreachable")
	}
	return exchange.NewWebsocket(
		exchange.ChannelId{channel, exchange.Binance},
		endpoint,
		&Conv{false}).Subscribe()
}

func (_ *Api) SubscribeCombined(pairs []exchange.CoinPair, channel exchange.Channel) error {
	endpoint := combinedBaseURL
	for _, p := range pairs {
		switch channel {
		case exchange.Candlestick:
			endpoint += fmt.Sprintf("%s@kline_%s/", internal.MakeSymbol(p), "1m")
		case exchange.Trade:
			endpoint += fmt.Sprintf("%s@trade/", internal.MakeSymbol(p))
		default:
			panic("unreachable")
		}
	}
	endpoint = endpoint[:len(endpoint)-1]
	return exchange.NewWebsocket(
		exchange.ChannelId{channel, exchange.Binance},
		endpoint,
		&Conv{true}).Subscribe()

	/*return &exchange.UnsupportedError{
	Message:
	fmt.Sprintf(
		"Binance does not suppord combined subscribe on channle %v",
		channel)}*/
}

func (_ *Api) IsSupported(pair exchange.CoinPair) bool {
	for _, i := range pair {
		if _, ok := internal.Coins[i]; !ok {
			return false
		}
	}
	return true
}

func (api *Api) FilterSupported(pairs []exchange.CoinPair) []exchange.CoinPair {
	var r []exchange.CoinPair
	for _, p := range pairs {
		if api.IsSupported(p) {
			r = append(r, p)
		}
	}
	return r
}

func (_ *Api) Unsubscribe(exchange.Channel) error {
	return nil
}

func (_ *Api) UnsubscribeAll() error {
	return nil
}

func (cv *Conv) Conv(c exchange.ChannelId, m []byte) (interface{}, error) {
	switch c.Channel {
	case exchange.Candlestick:
		if msg, err := internal.CandlelstickDecode(cv.combined, m); err != nil {
			return nil, err
		} else {
			return msg, nil
		}
	case exchange.Trade:
		if msg, err := internal.TradeDecode(cv.combined, m); err != nil {
			return nil, err
		} else {
			return msg, nil
		}
	}
	return nil, nil
}
