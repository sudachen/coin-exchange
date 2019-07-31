package tests

import (
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory"
	"gotest.tools/assert"
	"testing"
)

func Test_Binance(t *testing.T) {
	api := apifactory.Get(exchange.Binance)
	assert.Assert(t, api != nil)
	b := api.IsSupported(exchange.CoinPair{exchange.BTC, exchange.ETC})
	assert.Assert(t, b)
}

func Test_Huobi(t *testing.T) {
	api := apifactory.Get(exchange.Huobi)
	assert.Assert(t, api != nil)
	b := api.IsSupported(exchange.CoinPair{exchange.BTC, exchange.ETC})
	assert.Assert(t, b)
}

func Test_Okex(t *testing.T) {
	api := apifactory.Get(exchange.Okex)
	assert.Assert(t, api != nil)
	b := api.IsSupported(exchange.CoinPair{exchange.BTC, exchange.ETC})
	assert.Assert(t, b)
}
