package exchange

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type CoinType byte

const (
	NoCoin CoinType = iota
	USD             // Fiat
	BTC             // Bitcoin
	ETH             // Ethereum
	XRP             // Ripple
	LTC             // Litecoin
	BCH             // Bitcoin Cash
	BNB				// Binance Coin
	EOS				// EOS
)

func (c CoinType) String() string {
	switch c {
	case USD:
		return "USD"
	case BTC:
		return "BTC"
	case ETH:
		return "ETH"
	case XRP:
		return "XRP"
	case LTC:
		return "LTC"
	case BCH:
		return "BCH"
	case BNB:
		return "BNB"
	case EOS:
		return "EOS"
	}
	panic("unreachable")
}

func CoinFromString(s string) (CoinType, error) {
	switch strings.ToUpper(s) {
	case "USD":
		return USD, nil
	case "BTC":
		return BTC, nil
	case "ETH":
		return ETH, nil
	case "XRP":
		return XRP, nil
	case "LTC":
		return LTC, nil
	case "BCH":
		return BCH, nil
	case "BNB":
		return BNB, nil
	case "EOS":
		return EOS, nil
	default:
		return NoCoin, fmt.Errorf("unknown coin %v", s)
	}
}

type CoinPair [2]CoinType

var NoPair = CoinPair{NoCoin, NoCoin}

func (c *CoinType) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag != "!!str" {
		return fmt.Errorf("can't decode coin")
	}

	if v, err := CoinFromString(value.Value); err != nil {
		return err
	} else {
		*c = v
	}

	return nil
}
