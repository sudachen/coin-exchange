package exchange

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type Exchange byte

const (
	NoExchange Exchange = iota
	Binance
)

func (e Exchange) String() string {
	switch e {
	case Binance:
		return "Binance"
	}
	panic("unreachable")
}

func ExchangeFromString(s string) (Exchange, error) {
	switch strings.Title(s) {
	case "Binance":
		return Binance, nil
	default:
		return NoExchange, fmt.Errorf("unknown exchange platform %v", s)
	}
}

type Api interface {
	Subscribe(pair CoinPair, channel Channel) error
	SubscribeCombined(pairs []CoinPair, channel Channel) error
	IsSupported(pair CoinPair) bool
	FilterSupported(pairs []CoinPair) []CoinPair
	Unsubscribe(channel Channel) error
	UnsubscribeAll() error
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
