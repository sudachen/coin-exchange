package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"time"
)

type Candlestick struct {
	// Id int64 `json:"id"`
	Amount float32 `json:"amount"`
	Count  int32   `json:"count"`
	Open   float32 `json:"open"`
	Close  float32 `json:"close"`
	Low    float32 `json:"low"`
	High   float32 `json:"high"`
	Vol    float32 `json:"vol"`
}

type CandlestickCombined struct {
	Ch   string      `json:"ch"`
	Ts   int64       `json:"ts"`
	Data Candlestick `json:"tick"`
}

func CandlestickDecode(m []byte) ([]*message.Candlestick, error) {

	c := CandlestickCombined{}
	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	sym := c.Ch[:len(c.Ch)-11][7:]
	pair := SymbolToPair(sym)
	if pair == nil {
		return nil, fmt.Errorf("unsupported symbol '%v' in Candlestick message", sym)
	}

	mesg := &message.Candlestick{
		Origin:    exchange.Huobi,
		Pair:      *pair,
		Kline:	   message.Kline{
			Timestamp: time.Unix(c.Ts/1000, (c.Ts%1000)*1000000),
			TradeNum:  c.Data.Count,
			Open:      c.Data.Open,
			Close:     c.Data.Close,
			High:      c.Data.High,
			Low:       c.Data.Low,
			Volume:    c.Data.Vol,
			Interval:  1,
		},
	}

	return []*message.Candlestick{mesg}, nil
}
