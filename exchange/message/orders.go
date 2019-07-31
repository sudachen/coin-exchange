package message

import (
	"github.com/sudachen/coin-exchange/exchange"
	"strconv"
	"time"
)

type OrderValue struct {
	Price float32
	Qty   float32
}

type DepthAgg struct {
	NormCenter float32
	Center float32
	Volume float32
	Value  float32
}

type Orders struct {
	Origin exchange.Exchange
	Pair   exchange.CoinPair
	Bids   []OrderValue
	Asks   []OrderValue
	//AggBids   DepthAgg
	//AggAsks   DepthAgg
	Timestamp time.Time
}

func MakeDepthValues(a [][]string) []OrderValue {
	r := make([]OrderValue, len(a))

	cv := func(s string) float32 {
		if f, err := strconv.ParseFloat(s, 32); err != nil {
			return 0
		} else {
			return float32(f)
		}
	}

	for i, v := range a {
		r[i] = OrderValue{cv(v[0]), cv(v[1])}
	}

	return r
}

func CalcDepthAgg(asks []OrderValue, bids []OrderValue) (aAgg DepthAgg, bAgg DepthAgg) {
	agg := [2]DepthAgg{}
	ord := [][]OrderValue{asks,bids}
	qtys := [2]float32{}

	for i:=0; i<2; i++ {
		for _,v := range ord[i] {
			qtys[i] += v.Qty
			agg[i].Value += v.Qty * v.Price
		}
	}

	Qty := qtys[0]
	if Qty > qtys[1] {
		Qty = qtys[1]
	}

	for i:=0; i<2; i++ {
		agg[i].NormCenter = center2(total2(Qty,ord[i]),ord[i])/Qty
		agg[i].Center = center2(total2(qtys[i],ord[i]),ord[i])/qtys[i]
		agg[i].Volume = qtys[i]
	}

	//fmt.Println(agg,asks[:3],bids[:3])

	return agg[0], agg[1]
}


func total2(Qty float32, ord []OrderValue) float32 {
	total := float32(0)
	qty := float32(0)
	loop1:for _, v := range ord {
		if Qty >= qty+v.Qty {
			total += v.Qty * v.Price
			qty += v.Qty
		} else if Qty > qty {
			total += (Qty - qty) * v.Price
			break loop1
		} else {
			break loop1
		}
	}
	return total/2
}

func center2(total2 float32, ord []OrderValue) float32 {
	qty := float32(0)
	acc := float32(0)
	loop1:for _,v := range ord {
		if acc < total2 {
			mas := v.Qty * v.Price
			if acc + mas <= total2 {
				acc += mas
				qty += v.Qty
			} else {
				qty += (total2 - acc)/v.Price
				break loop1
			}
		} else {
			break loop1
		}
	}
	return qty
}

