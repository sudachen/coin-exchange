package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strconv"
	"time"
)

type Candlestick struct {
	Kline      []string `json:"candle"`
	Instrument string   `json:"instrument_id"`
}

type CandlestickCombined struct {
	Stream string        `json:"table"`
	Data   []Candlestick `json:"data"`
}

func CandlestickDecode(m []byte) ([]*message.Candlestick, error) {

	var r []*message.Candlestick
	c := CandlestickCombined{}

	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	for _, e := range c.Data {
		pair := SymbolToPair(e.Instrument)
		if pair == nil {
			return nil, fmt.Errorf("unsupported symbol '%v' in Candlestick message", e.Instrument)
		}

		cv := func(s string) float32 {
			if f, err := strconv.ParseFloat(s, 32); err != nil {
				return 0
			} else {
				return float32(f)
			}
		}

		theTime, _ := time.Parse(tmLayout, e.Kline[0])

		mesg := &message.Candlestick{
			Origin:       exchange.Okex,
			Pair:         *pair,
			StartTime:    theTime,
			EndTime:      theTime,
			FirstTradeId: 0,
			LastTradeId:  0,
			TradeNum:     0,
		}

		mesg.Open = cv(e.Kline[1])
		mesg.Close = cv(e.Kline[4])
		mesg.High = cv(e.Kline[2])
		mesg.Low = cv(e.Kline[3])
		mesg.Volume = cv(e.Kline[5])

		mesg.Interval = 1 //min
		r = append(r, mesg)
	}

	return r, nil
}
