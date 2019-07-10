package rawmesg

import (
	"encoding/json"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/message"
	"strconv"
	"time"
)

type Kline struct {
	StartTime 		int64	`json:"t"`
	EndTime 		int64	`json:"T"`
	Symbol			string	`json:"s"`
	Interval 		string  `json:"i"`
	FirstTradeId  	int64	`json:"f"`
	LastTradeId  	int64	`json:"L"`
	Open 			string  `json:"o"`
	Close 			string  `json:"c"`
	High    		string  `json:"h"`
	Low    			string  `json:"l"`
	Volume    		string  `json:"v"`
	TradeNum        int32   `json:"n"`
	IsFinal 		bool	`json:"x"`
	QuteVolume		string	`json:"q"`
	ActiveBuyVolume string 	`json:"V"`

	ActiveBuyQuoteVolume string `json:"Q"`

	IgnoreMe 		string 	`json:"B"`
}

type Candlelstick struct {
	EventType 		string 	`json:"e"`
	EventTime		int64	`json:"E"`
	Symbol 			string  `json:"s"`
	Kline			Kline   `json:"k"`
}

func CandlelstickDecode(sid exchange.StreamId, m []byte) (*message.Candlestick, error) {
	e := Candlelstick{}
	if err := json.Unmarshal(m, &e); err != nil {
		return nil, err
	}

	mesg := &message.Candlestick{
		Origin: sid,
		StartTime: time.Unix(e.Kline.StartTime,0),
		EndTime: time.Unix(e.Kline.EndTime,0),
		FirstTradeId: e.Kline.FirstTradeId,
		LastTradeId: e.Kline.LastTradeId,
		TradeNum: e.Kline.TradeNum,
	}

	cv := func(s string) float32 {
		if f, err := strconv.ParseFloat(s, 32); err != nil {
			return 0
		} else {
			return float32(f)
		}}

	mesg.Open = cv(e.Kline.Open)
	mesg.Close = cv(e.Kline.Close)
	mesg.High = cv(e.Kline.High)
	mesg.Low = cv(e.Kline.Low)
	mesg.Volume = cv(e.Kline.Volume)

	mesg.Interval = 1 //min

	return mesg, nil;
}
