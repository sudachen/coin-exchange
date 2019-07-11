package exchange

type Channel byte

const (
	Trade Channel = iota
	Candlestick
)

const MsgQueueLength = 61
const ErrQueueLength = 13

type StreamId struct {
	Pair     CoinPair
	Channel  Channel
	Exchange Exchange
}

var Collector = &StreamMachine{
	make(chan interface{}, MsgQueueLength),
	make(chan WebsocketError, ErrQueueLength),
	make(map[StreamId]*Websocket)}

type StreamMachine struct {
	MsgStream chan interface{}
	ec        chan WebsocketError
	streams   map[StreamId]*Websocket
}

func (m *StreamMachine) Subscribe(ws *Websocket) error {
	return ws.Connect(m.MsgStream, m.ec)
}

func (m *StreamMachine) Next() interface{} {
	return <-m.MsgStream
}
