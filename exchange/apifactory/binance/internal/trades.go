package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"time"
)

type Trade struct {
	//EventType      string `json:"e"`
	//EventTime      int64  `json:"E"`
	Symbol string `json:"s"`
	//TradeId        int64  `json:"t"`
	Price float32 `json:"p,string"`
	Qty   float32 `json:"q,string"`
	//BuyerOrderId   int64  `json:"b"`
	//SellerOrderId  int64  `json:"a"`
	TradeOrderTime int64 `json:"T"`
	IsBuyerMarket  bool  `json:"m"`
	//IgnoreMe       bool   `json:"M"`
}

type TradeCombined struct {
	Stream string `json:"stream"`
	Data   *Trade `json:"data"`
}

func TradeDecode(m []byte) (*message.Trade, error) {

	e := Trade{}
	c := TradeCombined{Data: &e}
	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	pair := SymbolToPair(e.Symbol)
	if pair == nil {
		return nil, fmt.Errorf("unsupported symbol '%v' in Trade message", e.Symbol)
	}

	mesg := &message.Trade{
		Origin: exchange.Binance,
		Pair:   *pair,
		Value: message.TradeValue{
			Sell:      !e.IsBuyerMarket,
			Timestamp: time.Unix(e.TradeOrderTime/1000, (e.TradeOrderTime%1000)*1000000),
			Price:     e.Price,
			Qty:       e.Qty,
		},
	}

	return mesg, nil
}

type AggTrade struct {
	Price         float32 `json:"p,string"`
	Qty           float32 `json:"q,string"`
	Timestamp     int64   `json:"T"`
	IsBuyerMarket bool    `json:"m"`
}

type AggTrades []AggTrade

func (a AggTrades) ToValues() []message.TradeValue {
	r := make([]message.TradeValue, len(a))
	for i, v := range a {
		r[i] = message.TradeValue{
			v.Price,
			v.Qty,
			!v.IsBuyerMarket,
			time.Unix(v.Timestamp/1000, (v.Timestamp%1000)*1000000)}
	}
	return r
}
