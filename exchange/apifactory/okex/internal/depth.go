package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strconv"
)

type Depth struct {
	Timestamp  string     `json:"timestamp"`
	Instrument string     `json:"instrument_id"`
	Asks       [][]string `json:"asks"`
	Bids       [][]string `json:"bids"`
}

type DepthCombined struct {
	Stream string  `json:"table"`
	Data   []Depth `json:"data"`
}

func DepthDecode(m []byte) ([]*message.Depth, error) {
	var r []*message.Depth
	c := DepthCombined{}

	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	for _, e := range c.Data {
		pair := SymbolToPair(e.Instrument)
		if pair == nil {
			return nil, fmt.Errorf("unsupported symbol '%v' in Depth message", e.Instrument)
		}

		mesg := &message.Depth{
			Origin: exchange.Okex,
			Pair:   *pair,
			Bids:   make([]message.DepthValue, len(e.Bids)),
			Asks:   make([]message.DepthValue, len(e.Asks)),
		}

		cv := func(s string) float32 {
			if f, err := strconv.ParseFloat(s, 32); err != nil {
				return 0
			} else {
				return float32(f)
			}
		}

		for i, v := range e.Bids {
			mesg.Bids[i] = message.DepthValue{cv(v[0]), cv(v[1])}
		}
		for i, v := range e.Asks {
			mesg.Asks[i] = message.DepthValue{cv(v[0]), cv(v[1])}
		}

		r = append(r, mesg)
	}

	return r, nil
}
