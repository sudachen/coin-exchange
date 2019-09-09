package channel

import (
	"fmt"
	"strconv"
	"strings"
)

type Channel int32

const (
	NoChannel Channel = iota
	Candlestick
	Trade
	Depth
)

const AnotherChannel Channel = Depth + 1

func (c Channel) String() string {
	switch c {
	case NoChannel:
		return "no-channel"
	case Trade:
		return "Trade"
	case Candlestick:
		return "Candlestick"
	case Depth:
		return "Depth"
	default:
		return fmt.Sprintf("Channel-%d", c)
	}
}

func FromString(s string) Channel {
	switch strings.ToLower(s) {
	case "no-channel": return NoChannel
	case "trade": return Trade
	case "candlestick": return Candlestick
	case "depth": return Depth
	default:
		if strings.Index(strings.ToLower(s),"channel-") == 0 {
			v, err := strconv.ParseInt(s[8:],10,32)
			if err == nil {
				return Channel(v)
			}
		}
	}
	return NoChannel
}
