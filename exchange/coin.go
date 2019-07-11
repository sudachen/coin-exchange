package exchange

import (
	"fmt"
	"strings"
)

type CoinType byte

const (
	USD CoinType = iota
	BTC
	ETH
)

func (c CoinType) String() string {
	switch c {
	case USD:
		return "USD"
	case BTC:
		return "BTC"
	case ETH:
		return "ETH"
	}
	panic("unreachable")
}

func FromString(s string) CoinType {
	switch strings.ToUpper(s) {
	case "USD":
		return USD
	case "BTC":
		return BTC
	case "ETC":
		return ETH
	}
	panic(fmt.Sprintf("unknown coin %v", s))
}

type CoinPair [2]CoinType
