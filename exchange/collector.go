package exchange


type Channel byte

const (
	NoChannel Channel = iota
	Trade
	Candlestick
	Depth
)

func (c Channel) String() string {
	switch c {
	case Trade:
		return "Trade"
	case Candlestick:
		return "Candlestick"
	case Depth:
		return "Depth"
	default:
		panic("unreachable")
	}
}

const MsgQueueLength = 61

type collector struct {Messages chan interface{}}

var Collector = &collector{
	make(chan interface{}, MsgQueueLength),
}

func (m *collector) Next() interface{} {
	return <-m.Messages
}

