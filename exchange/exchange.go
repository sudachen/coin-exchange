package exchange

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

type Exchange byte

const (
	NoExchange Exchange = iota
	Binance
	Okex
)

func (e Exchange) String() string {
	switch e {
	case Binance:
		return "Binance"
	case Okex:
		return "Okex"
	}
	panic("unreachable")
}

func ExchangeFromString(s string) (Exchange, error) {
	switch strings.Title(s) {
	case "Binance":
		return Binance, nil
	case "Okex":
		return Okex, nil
	default:
		return NoExchange, fmt.Errorf("unknown exchange platform %v", s)
	}
}

type Api interface {
	Subscribe(pairs []CoinPair, channels []Channel) error
	IsSupported(pair CoinPair) bool
	FilterSupported(pairs []CoinPair) []CoinPair
	UnsubscribeAll(duration time.Duration) error
}

func (e *Exchange) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag != "!!str" {
		return fmt.Errorf("can't decode exchange")
	}

	if v, err := ExchangeFromString(value.Value); err != nil {
		return err
	} else {
		*e = v
	}

	return nil
}

type UnsupportedError struct {
	Message string
}

func (e *UnsupportedError) Error() string {
	return e.Message
}
