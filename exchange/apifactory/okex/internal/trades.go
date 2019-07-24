package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strconv"
	"time"
)

type Trade struct {
	Price      string `json:"price"`
	Side       string `json:"side"`
	Size       string `json:"size"`
	Timestamp  string `json:"timestamp"`
	TradeId    string `json:"trade_id"`
	Instrument string `json:"instrument_id"`
}

type TradeCombined struct {
	Stream string  `json:"table"`
	Data   []Trade `json:"data"`
}

func TradeDecode(m []byte) ([]*message.Trade, error) {

	var r []*message.Trade
	c := TradeCombined{}

	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	for _, e := range c.Data {
		pair := SymbolToPair(e.Instrument)
		if pair == nil {
			return nil, fmt.Errorf("unsupported symbol '%v' in Trade message", e.Instrument)
		}

		cv := func(s string) float32 {
			if f, err := strconv.ParseFloat(s, 32); err != nil {
				return 0
			} else {
				return float32(f)
			}
		}

		theTime, _ := time.Parse(tmLayout, e.Timestamp)

		mesg := &message.Trade{
			Origin: exchange.Okex,
			Pair:   *pair,
			Value: message.TradeValue{
				Sell:      e.Side == "sell",
				Timestamp: theTime,
				Price:     cv(e.Price),
				Qty:       cv(e.Size),
			},
		}

		r = append(r, mesg)
	}

	return r, nil
}
