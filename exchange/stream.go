package exchange

import "sync"

type Channel byte

const (
	Trade Channel = iota
	Candlestick
)

func (c Channel) String() string {
	switch c {
	case Trade:
		return "Trade"
	case Candlestick:
		return "Candlestick"
	default:
		panic("unreachable")
	}
}

const MsgQueueLength = 61
const ErrQueueLength = 13

type ChannelId struct {
	Channel  Channel
	Exchange Exchange
}

var Collector = &StreamMachine{
	make(chan interface{}, MsgQueueLength),
	make(chan WebsocketError, ErrQueueLength),
	make(map[ChannelId][]*Websocket),
	&sync.Mutex{},
}

type StreamMachine struct {
	MsgStream chan interface{}
	ec        chan WebsocketError
	streams   map[ChannelId][]*Websocket
	mutex     *sync.Mutex
}

func (m *StreamMachine) Subscribe(ws *Websocket) error {
	m.mutex.Lock()
	if l, ok := m.streams[ws.Cid]; ok {
		exists := false
		for _, k := range l {
			if k == ws {
				exists = true
				break
			}
		}
		if !exists {
			m.streams[ws.Cid] = append(l, ws)
		}
	} else {
		m.streams[ws.Cid] = []*Websocket{ws}
	}
	m.mutex.Unlock()
	return ws.Connect(m.MsgStream, m.ec)
}

func (m *StreamMachine) Next() interface{} {
	return <-m.MsgStream
}
