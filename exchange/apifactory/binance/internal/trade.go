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

type TradeCombined struct {
	Stream string `json:"stream"`
	Data   *Trade `json:"data"`
}

func TradeDecode(combined bool, m []byte) (*message.Trade, error) {

	e := Trade{}
	if combined {
		c := TradeCombined{Data: &e}
		if err := json.Unmarshal(m, &c); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(m, &e); err != nil {
			return nil, err
		}
	}

	pair := SymbolToPair(e.Symbol)
	if pair == nil {
		return nil, fmt.Errorf("unsupported symbol '%v' in Trade message", e.Symbol)
	}

	mesg := &message.Trade{
		Origin:         exchange.Binance,
		Pair:           *pair,
		TradeId:        e.TradeId,
		BuyerOrderId:   e.BuyerOrderId,
		SellerOrderId:  e.SellerOrderId,
		TradeOrderTime: time.Unix(e.TradeOrderTime/1000,(e.TradeOrderTime%1000)*1000000),
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
