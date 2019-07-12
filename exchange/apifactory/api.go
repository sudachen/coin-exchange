package apifactory

import (
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/binance/api"
)

func Get(ex exchange.Exchange) exchange.Api {
	switch ex {
	case exchange.Binance:
		return &api.Api{}
	default:
		panic(fmt.Sprintf("unknown exchange %v", ex.String()))
	}
}
