package apifactory

import (
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	binance "github.com/sudachen/coin-exchange/exchange/apifactory/binance/api"
	huobi "github.com/sudachen/coin-exchange/exchange/apifactory/huobi/api"
	okex "github.com/sudachen/coin-exchange/exchange/apifactory/okex/api"
	"github.com/sudachen/coin-exchange/exchange/message"
	"sync"
	"time"
)

var apis = make(map[exchange.Exchange]message.Api)

func Get(ex exchange.Exchange) message.Api {
	api, ok := apis[ex]
	if !ok {
		switch ex {
		case exchange.Binance:
			api = binance.New()
		case exchange.Okex:
			api = okex.New()
		case exchange.Huobi:
			api = huobi.New()
		default:
			panic(fmt.Sprintf("unknown exchange %v", ex.String()))
		}
		apis[ex] = api
	}
	return api
}

func UnsubscribeAll(timeout time.Duration, wg *sync.WaitGroup) {
	wgg := wg
	if wgg == nil {
		wgg = &sync.WaitGroup{}
	}
	for _, api := range apis {
		api.UnsubscribeAll(timeout, wgg)
	}
	if wg == nil {
		wgg.Wait()
	}
}

