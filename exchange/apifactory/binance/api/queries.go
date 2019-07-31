package api

import (
	"encoding/json"
	"fmt"
	"github.com/sudachen/coin-exchange/exchange"
	"github.com/sudachen/coin-exchange/exchange/apifactory/binance/internal"
	"github.com/sudachen/coin-exchange/exchange/message"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const apiUrl = "https://api.binance.com/api/v1/"

type queries struct {
	*api
	Pair exchange.CoinPair
	*http.Client
}

func (a *api) Queries(pair exchange.CoinPair) (message.QueryApi, error) {
	a.m.Lock()
	excluded := a.exclude[pair]
	a.m.Unlock()
	if internal.Coins[pair[0]] && internal.Coins[pair[1]] && !excluded {
		return &queries{
			a,
			pair,
			&http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 6,
					TLSHandshakeTimeout: 3 * time.Second,
				},
			},
		}, nil
	} else {
		return nil, &exchange.UnsupportedPair{exchange.Binance, pair}
	}
}

func (q *queries) QueryDepth(l message.Limit) (*message.Orders, error) {
	var depth internal.Depth
	var count int
	switch l {
	case message.LimitMin:
		count = 5
	case message.LimitLow:
		count = 20
	case message.LimitNorm:
		count = 100
	case message.LimitHigh:
		count = 500
	case message.LimitMax:
		count = 1000
	}
	if err := q.query("depth", count, "", &depth); err != nil {
		return nil, err
	} else {
		return &message.Orders{
			Origin:    exchange.Binance,
			Pair:      q.Pair,
			Timestamp: time.Now(),
			Bids:      message.MakeDepthValues(depth.Bids),
			Asks:      message.MakeDepthValues(depth.Asks),
		}, nil
	}
}

func (q *queries) QueryTrades(l message.Limit) (*message.Trades, error) {
	return q.queryTrades(l, false)
}

func (q *queries) QueryAggTrades(l message.Limit) (*message.Trades, error) {
	return q.queryTrades(l, true)
}

func (q *queries) queryTrades(l message.Limit, agg bool) (*message.Trades, error) {
	var count int
	switch l {
	case message.LimitMin:
		count = 5
	case message.LimitLow:
		count = 20
	case message.LimitNorm:
		count = 100
	case message.LimitHigh:
		count = 500
	case message.LimitMax:
		count = 1000
	}

	var val []message.TradeValue

	if agg {
		t := internal.AggTrades{}
		if err := q.query("aggTrades", count, "", &t); err != nil {
			return nil, err
		}
		val = t.ToValues()
	} else {
		t := internal.HistTrades{}
		if err := q.query("trades", count, "", &t); err != nil {
			return nil, err
		}
		val = t.ToValues()
	}

	return &message.Trades{
		Origin: exchange.Binance,
		Pair:   q.Pair,
		Values: val,
	}, nil
}

func (q *queries) QueryCandlesticks(interval int, count int) (*message.Candlesticks, error) {
	var i string
	var minutes int32
	if interval >= 60*24 {
		i = "1d"
		minutes = 60 * 24
	} else if interval >= 60 {
		i = "1h"
		minutes = 60
	} else if interval >= 30 {
		i = "30m"
		minutes = 30
	} else if interval >= 15 {
		i = "15m"
		minutes = 15
	} else if interval >= 5 {
		i = "5m"
		minutes = 5
	} else {
		i = "1m"
		minutes = 1
	}

	var klines []Kline
	if err := q.query("klines", count, i, &klines); err != nil {
		return nil, err
	} else {
		r := &message.Candlesticks{}
		r.Pair = q.Pair
		r.Origin = exchange.Binance
		r.Klines = make([]message.Kline, len(klines))
		for i, v := range klines {
			r.Klines[i] = v.Kline
			r.Klines[i].Interval = minutes
		}
		return r, nil
	}
}

func (q *queries) query(api string, limit int, interval string, result interface{}) error {
	symbol := strings.ToUpper(internal.MakeSymbol(q.Pair))
	url := apiUrl + api + "?symbol=" + symbol
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	if interval != "" {
		url += "&interval=" + interval
	}

	//logger.Info(url)
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		return err
	} else {
		req.Header.Add("Accept", "application/json")

		if resp, err := q.Client.Do(req); err != nil {
			return err
		} else {
			defer resp.Body.Close()
			if body, err := ioutil.ReadAll(resp.Body); err != nil {
				return err
			} else {
				if resp.StatusCode >= 400 {
					apiErr := &apiError{Pair: q.Pair, Origin: exchange.Binance}
					_ = json.Unmarshal(body, apiErr)
					if apiErr.Code == -1121 { // invalid symbol
						q.api.m.Lock()
						q.api.exclude[q.Pair] = true
						q.api.m.Unlock()
						return &exchange.UnsupportedPair{exchange.Binance, q.Pair}
					}
					return apiErr
				} else {
					if resp != nil {
						err := json.Unmarshal(body, result)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}
		}
	}
}

type Kline struct {
	message.Kline
}

func (k *Kline) UnmarshalJSON(b []byte) error {
	var s [11]interface{}
	err := json.Unmarshal(b, &s)

	if err != nil {
		return err
	}

	k.Timestamp = time.Unix(int64(s[0].(float64))/1000, 0)

	cv := func(i int) float32 {
		if f, err := strconv.ParseFloat(s[1].(string), 64); err != nil {
			return 0
		} else {
			return float32(f)
		}
	}

	k.Open = cv(1)
	k.High = cv(2)
	k.Low = cv(3)
	k.Close = cv(4)
	k.Volume = cv(5)

	k.TradeNum = int32(s[8].(float64))

	return nil
}

type apiError struct {
	Origin  exchange.Exchange `json:"-"`
	Pair    exchange.CoinPair `json:"-"`
	Code    int64             `json:"code"`
	Message string            `json:"msg"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("apiError{%v|%v on %v:%v}", e.Code, e.Message, e.Origin.String(), e.Pair.String())
}
