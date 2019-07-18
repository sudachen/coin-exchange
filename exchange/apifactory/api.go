package apifactory

import (
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	binance "github.com/sudachen/coin-exchange/exchange/apifactory/binance/api"
	okex "github.com/sudachen/coin-exchange/exchange/apifactory/okex/api"
)

var apis = make(map[exchange.Exchange]exchange.Api)

func Get(ex exchange.Exchange) exchange.Api {
	api, ok := apis[ex]
	if !ok {
		switch ex {
		case exchange.Binance:
			api = binance.New()
		case exchange.Okex:
			api = okex.New()
		default:
			panic(fmt.Sprintf("unknown exchange %v", ex.String()))
		}
		apis[ex] = api
	}
	return api
}
