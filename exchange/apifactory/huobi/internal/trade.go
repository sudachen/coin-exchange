package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"time"
)

type Trade struct {
	Direction string  `json:"direction"`
	Price     float32 `json:"price"`
	Amount    float32 `json:"amount"`
	Ts        int64   `json:"ts"`
}

type TradeTick struct {
	//Id int64 `json:"id"`
	//Ts int64 `json:"ts"`
	Data []Trade `json:"data"`
}

type TradeCombined struct {
	Ch string `json:"ch"`
	//Ts int64  `json:"ts"`
	Tk TradeTick `json:"tick"`
}

func TradeDecode(m []byte) ([]*message.Trade, error) {

	var r []*message.Trade
	c := TradeCombined{}

	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	for _, e := range c.Tk.Data {
		sym := c.Ch[:len(c.Ch)-13][7:]
		pair := SymbolToPair(sym)
		if pair == nil {
			return nil, fmt.Errorf("unsupported symbol '%v' in Trade message", sym)
		}

		mesg := &message.Trade{
			Origin:    exchange.Huobi,
			Pair:      *pair,
			Sell:      e.Direction == "sell",
			Timestamp: time.Unix(e.Ts/1000, (e.Ts%1000)*1000000),
			Price:     e.Price,
			Qty:       e.Amount,
		}

		r = append(r, mesg)
	}

	return r, nil
}
