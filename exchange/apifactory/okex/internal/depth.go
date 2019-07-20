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

func DepthDecode(m []byte) ([]*message.Depth, error) {
	var r []*message.Depth
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

		mesg := &message.Depth{
			Origin:    exchange.Okex,
			Pair:      *pair,
			Timestamp: theTime,
			Bids:      message.MakeDepthValues(e.Bids),
			Asks:      message.MakeDepthValues(e.Asks),
		}

		//bdp := message.MakeDepthValues(e.Bids)
		//mesg.AggBids = message.CalcDepthAgg(bdp)
		//mesg.Bids = bdp
		//adp := message.MakeDepthValues(e.Asks)
		//mesg.AggAsks = message.CalcDepthAgg(adp)
		//mesg.Asks = adp

		r = append(r, mesg)
	}

	return r, nil
}
