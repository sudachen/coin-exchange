package internal

import (
	"github.com/sudachen/coin-exchange/exchange"
	"strings"
)

var Coins = map[exchange.CoinType]bool{
	exchange.USD: true,
	exchange.BTC: true,
	exchange.ETH: true,
	exchange.XRP: true,
	exchange.LTC: true,
	exchange.BCH: true,
	exchange.BNB: true,
	exchange.EOS: true,
}

var Pairs = make(map[string]exchange.CoinPair)

func init() {
	for c1, _ := range Coins {
		for c2, _ := range Coins {
			if c1 != c2 {
				pair := exchange.CoinPair{c1, c2}
				Pairs[strings.ToUpper(MakeSymbol(pair))] = pair
			}
		}
	}
}

func MakeSymbol(pair exchange.CoinPair) string {
	var ccs [2]string
	for i, cc := range pair {
		switch cc{
		case exchange.USD: ccs[i] = "USDT"
		case exchange.BCH: ccs[i] = "BCHABC"
		default: ccs[i] = cc.String()
		}
	}
	return strings.ToLower(ccs[0] + ccs[1])
}

func SymbolToPair(symbol string) *exchange.CoinPair {
	if p, ok := Pairs[symbol]; ok {
		return &p
	}
	return nil
}
