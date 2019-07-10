package exchange

import (
	"github.com/gorilla/websocket"
	"time"
)

const wsTimeout = time.Second * 60
const wsKeepalive = false

type MsgConv = func(StreamId,[]byte) (interface{}, error)

type Websocket struct {
	Sid      StreamId
	endpoint string
	msgConv  MsgConv
	cClose   chan struct{}
}

func NewWebsocket(sid StreamId, endpoint string, msgConv MsgConv) *Websocket {
	return &Websocket {
		sid,
		endpoint,
		msgConv,
		nil,
	}
}

func (p *Websocket) Close() error {
	if p.cClose != nil {
		p.cClose <- struct{}{}
	}
	return nil
}

type WebsocketError struct {
	Sid StreamId
	Err error
}

func (p *Websocket) worker(mc chan interface{}, ec chan WebsocketError) {
	p.cClose = make(chan struct{})
	defer func (){ close(p.cClose); p.cClose = nil }()

	c, _, err := websocket.DefaultDialer.Dial(p.endpoint, nil)
	if err != nil {
		ec <- WebsocketError{ p.Sid, err }
		return
	}

	if wsKeepalive {
		p.keepAlive(c, wsTimeout, ec)
	}

	for {
		select {
		case <- p.cClose:
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				ec <- WebsocketError{ p.Sid, err }
				return
			}
			m, err := p.msgConv(p.Sid, message)
			if err != nil {
				ec <- WebsocketError{ p.Sid, err }
				return
			}
			if m != nil {
				mc <- m
			}
		}
	}
}

func (p *Websocket) keepAlive(c *websocket.Conn, timeout time.Duration, ec chan WebsocketError) {

	switch p.Sid.Channel {
	case Candlestick, Trade: return // are updated quite often
	}

	ticker := time.NewTicker(timeout)

	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		for {
			deadline := time.Now().Add(10 * time.Second)
			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				ec <- WebsocketError{ p.Sid, err }
				return
			}
			<-ticker.C
			if time.Now().Sub(lastResponse) > timeout {
				_ = c.Close()
				return
			}
		}
	}()
}

func (p *Websocket) Subscribe() error {
	return Collector.Subscribe(p)
}

func (p *Websocket) Connect(mc chan interface{}, ec chan WebsocketError) error {
	go p.worker(mc,ec)
	return nil
}
