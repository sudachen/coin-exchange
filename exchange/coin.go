package exchange

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"sort"
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
	BNB             // Binance Coin
	EOS             // EOS
	ADA             // Cardano
	ETC             // Ethereum Classic
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
	case ADA:
		return "ADA"
	case ETC:
		return "ETC"
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
	case "ADA":
		return ADA, nil
	case "ETC":
		return ETC, nil
	default:
		return NoCoin, fmt.Errorf("unknown coin %v", s)
	}
}

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

func (c CoinType) MarshalYAML() (interface{}, error) {
	return c.String(), nil
}

type CoinPair [2]CoinType

var NoPair = CoinPair{NoCoin, NoCoin}

func (c CoinPair) String() string {
	return fmt.Sprintf("%s/%s", c[0].String(), c[1].String())
}

func (c CoinPair) AsInt() int32 {
	return int32(c[0])*256 + int32(c[1])
}

func (c *CoinPair) FromInt(a int32) *CoinPair {
	*c = CoinPair{CoinType(a / 256), CoinType(a % 256)}
	return c
}

func PairFromString(s string) (CoinPair, error) {
	cs := strings.Split(s, "/")
	if len(cs) != 2 {
		return CoinPair{}, fmt.Errorf("'%v' is not coin pair", s)
	}
	p := CoinPair{}
	var err error
	for i := 0; i < 2; i++ {
		if p[i], err = CoinFromString(cs[i]); err != nil {
			return CoinPair{}, err
		}
	}
	return p, nil
}

func (p *CoinPair) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag != "!!str" {
		return fmt.Errorf("can't decode coin pair")
	}

	if v, err := PairFromString(value.Value); err != nil {
		return err
	} else {
		*p = v
	}

	return nil
}

func (p CoinPair) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

type UnsupportedPair struct {
	Exchange
	CoinPair
}

func (e *UnsupportedPair) Error() string {
	return fmt.Sprintf("UnsupportedPair{%v:%v}", e.Exchange.String(), e.CoinPair.String())
}

func GetUniqueCoins(pairs []CoinPair) []CoinType {
	cSet := map[CoinType]bool{}
	for _, k := range pairs {
		cSet[k[0]] = true
		cSet[k[1]] = true
	}
	coins := make([]CoinType,0,len(cSet))
	for k := range cSet {
		coins = append(coins,k)
	}
	sort.Slice(coins,func(i,j int)bool{return coins[i]<coins[j]})
	return coins
}
