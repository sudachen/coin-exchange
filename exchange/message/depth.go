package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"strconv"
	"time"
)

type DepthValue struct {
	Price float32
	Qty   float32
}

type DepthAgg struct {
	Avg    float32
	Median float32
	Qty    float32
	Volume float32
}

type Depth struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair
	Bids   []DepthValue
	Asks   []DepthValue
	//AggBids   DepthAgg
	//AggAsks   DepthAgg
	Timestamp time.Time
}

func MakeDepthValues(a [][]string) []DepthValue {
	r := make([]DepthValue, len(a))

	cv := func(s string) float32 {
		if f, err := strconv.ParseFloat(s, 32); err != nil {
			return 0
		} else {
			return float32(f)
		}
	}

	for i, v := range a {
		r[i] = DepthValue{cv(v[0]), cv(v[1])}
	}

	return r
}

func CalcDepthAgg(dp []DepthValue) DepthAgg {
	agg := DepthAgg{}
	for _, v := range dp {
		agg.Volume += v.Price * v.Qty
		agg.Qty += v.Qty
	}
	a := float32(0)
	for _, v := range dp {
		if a < agg.Volume/2 {
			agg.Median = v.Price
			a += v.Price * v.Qty
		} else {
			break
		}
	}
	a = float32(0)
	for _, v := range dp {
		if a < agg.Qty/2 {
			agg.Avg = v.Price
			a += v.Qty
		} else {
			break
		}
	}
	return agg
}
