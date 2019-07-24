package channel

import "fmt"

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
