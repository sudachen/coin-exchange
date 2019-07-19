package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strings"
	"time"
)

/*
{
  "lastUpdateId": 160,  // Last update ID
  "bids": [             // Bids to be updated
    [
      "0.0024",         // Price level to be updated
      "10"              // Quantity
    ]
  ],
  "asks": [             // Asks to be updated
    [
      "0.0026",         // Price level to be updated
      "100"            // Quantity
    ]
  ]
}
*/

type Depth struct {
	LastUpdateId int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type DepthCombined struct {
	Stream string `json:"stream"`
	Data   *Depth `json:"data"`
}

func DepthDecode(m []byte) (*message.Depth, error) {
	e := Depth{}
	c := DepthCombined{Data: &e}
	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	symbol := strings.ToUpper(strings.TrimSuffix(c.Stream, DepthSuffix))
	pair := SymbolToPair(symbol)
	if pair == nil {
		return nil, fmt.Errorf("unsupported symbol '%v' in Depth message", symbol)
	}

	mesg := &message.Depth{
		Origin:    exchange.Binance,
		Pair:      *pair,
		Timestamp: time.Now(),
		Bids:      message.MakeDepthValues(e.Bids),
		Asks:      message.MakeDepthValues(e.Asks),
	}

	//bdp := message.MakeDepthValues(e.Bids)
	//mesg.AggBids = message.CalcDepthAgg(bdp)
	//mesg.Bids = bdp
	//adp := message.MakeDepthValues(e.Asks)
	//mesg.AggAsks = message.CalcDepthAgg(adp)
	//mesg.Asks = adp

	return mesg, nil
}
