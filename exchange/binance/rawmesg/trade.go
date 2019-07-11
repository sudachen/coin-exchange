package rawmesg

import (
	"encoding/json"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strconv"
	"time"
)

type Trade struct {
	EventType      string `json:"e"`
	EventTime      int64  `json:"E"`
	Symbol         string `json:"s"`
	TradeId        int64  `json:"t"`
	Price          string `json:"p"`
	Qty            string `json:"q"`
	BuyerOrderId   int64  `json:"b"`
	SellerOrderId  int64  `json:"a"`
	TradeOrderTime int64  `json:"T"`
	IsBuyerMarket  bool   `json:"m"`
	IgnoreMe       bool   `json:"M"`
}

func TradeDecode(sid exchange.StreamId, m []byte) (*message.Trade, error) {
	e := Trade{}
	if err := json.Unmarshal(m, &e); err != nil {
		return nil, err
	}

	mesg := &message.Trade{
		Origin:         sid,
		TradeId:        e.TradeId,
		BuyerOrderId:   e.BuyerOrderId,
		SellerOrderId:  e.SellerOrderId,
		TradeOrderTime: time.Unix(e.TradeOrderTime, 0),
	}

	cv := func(s string) float32 {
		if f, err := strconv.ParseFloat(s, 32); err != nil {
			return 0
		} else {
			return float32(f)
		}
	}

	mesg.Price = cv(e.Price)
	mesg.Qty = cv(e.Qty)

	return mesg, nil
}
