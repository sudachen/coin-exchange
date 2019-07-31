package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"time"
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

func DepthDecode(m []byte) ([]*message.Orders, error) {
	var r []*message.Orders
	c := DepthCombined{}

	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	//logger.Infof("%#v",c)

	for _, e := range c.Data {
		pair := SymbolToPair(e.Instrument)
		if pair == nil {
			return nil, fmt.Errorf("unsupported symbol '%v' in Depth message", e.Instrument)
		}

		theTime, _ := time.Parse(tmLayout, e.Timestamp)

		//logger.Infof("Okex depth length: %v, %v", len(e.Bids), len(e.Asks))

		mesg := &message.Orders{
			Origin:    exchange.Okex,
			Pair:      *pair,
			Timestamp: theTime,
			Bids:      message.MakeDepthValues(e.Bids),
			Asks:      message.MakeDepthValues(e.Asks),
		}

		r = append(r, mesg)
	}

	return r, nil
}
