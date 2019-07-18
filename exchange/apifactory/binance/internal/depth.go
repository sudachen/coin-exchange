package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strconv"
	"strings"
	"time"
)

/*{
  "e": "depthUpdate", // Event type
  "E": 123456789,     // Event time
  "s": "BNBBTC",      // Symbol
  "U": 157,           // First update ID in event
  "u": 160,           // Final update ID in event
  "b": [              // Bids to be updated
    [
      "0.0024",       // Price level to be updated
      "10"            // Quantity
    ]
  ],
  "a": [              // Asks to be updated
    [
      "0.0026",       // Price level to be updated
      "100"           // Quantity
    ]
  ]
}*/
/*
type Depth struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	Symbol        string     `json:"s"`
	FirstUpdateId int64      `json:"U"`
	LastUpdateId  int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

type DepthCombined struct {
	Stream string `json:"stream"`
	Data   *Depth `json:"data"`
}
*/
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

type Depth5 struct {
	LastUpdateId int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type Depth5Combined struct {
	Stream string  `json:"stream"`
	Data   *Depth5 `json:"data"`
}

func DepthDecode(m []byte) (*message.Depth, error) {
	e := Depth5{}
	c := Depth5Combined{Data: &e}
	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	symbol := strings.ToUpper(strings.TrimSuffix(c.Stream, "@depth5"))
	pair := SymbolToPair(symbol)
	if pair == nil {
		return nil, fmt.Errorf("unsupported symbol '%v' in Depth message", symbol)
	}

	mesg := &message.Depth{
		Origin:    exchange.Binance,
		Pair:      *pair,
		Timestamp: time.Now(),
		Bids:      make([]message.DepthValue, len(e.Bids)),
		Asks:      make([]message.DepthValue, len(e.Asks)),
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

	return mesg, nil
}
