package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"time"
)

type DepthTick struct {
	Asks [][]float32 `json:"asks"`
	Bids [][]float32 `json:"bids"`
}

type DepthCombined struct {
	Ch string    `json:"ch"`
	Ts int64     `json:"ts"`
	Tk DepthTick `json:"tick"`
}

func DepthDecode(m []byte) ([]*message.Orders, error) {

	c := DepthCombined{}
	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	sym := c.Ch[:len(c.Ch)-12][7:]
	pair := SymbolToPair(sym)
	if pair == nil {
		return nil, fmt.Errorf("unsupported symbol '%v' in Depth message")
	}

	mesg := &message.Orders{
		Origin:    exchange.Huobi,
		Pair:      *pair,
		Timestamp: time.Unix(c.Ts/1000, (c.Ts%1000)*1000000),
		//Bids:
	}

	//logger.Infof("Huobi depth length: %v, %v", len(c.Tk.Bids), len(c.Tk.Asks))

	bdp := make([]message.OrderValue, len(c.Tk.Bids))
	for i, v := range c.Tk.Bids {
		bdp[i] = message.OrderValue{v[0], v[1]}
	}
	//mesg.AggBids = message.CalcDepthAgg(bdp)
	mesg.Bids = bdp

	adp := make([]message.OrderValue, len(c.Tk.Asks))
	for i, v := range c.Tk.Asks {
		adp[i] = message.OrderValue{v[0], v[1]}
	}
	//mesg.AggAsks = message.CalcDepthAgg(adp)
	mesg.Asks = adp

	//logger.Infof("asks: %d, bids: %d",len(mesg.Asks),len(mesg.Bids))

	return []*message.Orders{mesg}, nil
}
