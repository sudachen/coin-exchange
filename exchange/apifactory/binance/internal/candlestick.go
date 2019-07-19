package internal

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strconv"
	"time"
)

type Kline struct {
	StartTime       int64  `json:"t"`
	EndTime         int64  `json:"T"`
	Symbol          string `json:"s"`
	Interval        string `json:"i"`
	FirstTradeId    int64  `json:"f"`
	LastTradeId     int64  `json:"L"`
	Open            string `json:"o"`
	Close           string `json:"c"`
	High            string `json:"h"`
	Low             string `json:"l"`
	Volume          string `json:"v"`
	TradeNum        int32  `json:"n"`
	IsFinal         bool   `json:"x"`
	QuteVolume      string `json:"q"`
	ActiveBuyVolume string `json:"V"`

	ActiveBuyQuoteVolume string `json:"Q"`

	IgnoreMe string `json:"B"`
}

type Candlestick struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     Kline  `json:"k"`
}

type CandlestickCombined struct {
	Stream string       `json:"stream"`
	Data   *Candlestick `json:"data"`
}

func CandlestickDecode(m []byte) (*message.Candlestick, error) {

	e := Candlestick{}
	c := CandlestickCombined{Data: &e}

	if err := json.Unmarshal(m, &c); err != nil {
		return nil, err
	}

	pair := SymbolToPair(e.Symbol)
	if pair == nil {
		return nil, fmt.Errorf("unsupported symbol '%v' in Candlestick message", e.Symbol)
	}

	mesg := &message.Candlestick{
		Origin:    exchange.Binance,
		Pair:      *pair,
		Timestamp: time.Unix(e.Kline.EndTime/1000, (e.Kline.EndTime%1000)*1000000),
		TradeNum:  e.Kline.TradeNum,
	}

	cv := func(s string) float32 {
		if f, err := strconv.ParseFloat(s, 32); err != nil {
			return 0
		} else {
			return float32(f)
		}
	}

	mesg.Open = cv(e.Kline.Open)
	mesg.Close = cv(e.Kline.Close)
	mesg.High = cv(e.Kline.High)
	mesg.Low = cv(e.Kline.Low)
	mesg.Volume = cv(e.Kline.Volume)

	mesg.Interval = 1 //min

	return mesg, nil
}
