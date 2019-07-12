package exchange

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

const wsTimeout = time.Second * 60
const wsKeepalive = false

type MsgConv interface {
	Conv(ChannelId, []byte) (interface{}, error)
}

type Websocket struct {
	Cid      ChannelId
	endpoint string
	msgConv  MsgConv
	cClose   chan struct{}
}

func NewWebsocket(cid ChannelId, endpoint string, msgConv MsgConv) *Websocket {
	return &Websocket{
		cid,
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
	Ws  *Websocket
	Err error
}

func (ws *Websocket) worker(mc chan interface{}, ec chan WebsocketError) {
	ws.cClose = make(chan struct{})
	defer func() { close(ws.cClose); ws.cClose = nil }()

	c, _, err := websocket.DefaultDialer.Dial(ws.endpoint, nil)
	if err != nil {
		ec <- WebsocketError{ws, err}
		return
	}

	if wsKeepalive {
		ws.keepAlive(c, wsTimeout, ec)
	}

	for {
		select {
		case <-ws.cClose:
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println(err)
				ec <- WebsocketError{ws, err}
				return
			}
			m, err := ws.msgConv.Conv(ws.Cid, message)
			if err != nil {
				fmt.Println(err)
				ec <- WebsocketError{ws, err}
				return
			}
			if m != nil {
				mc <- m
			}
		}
	}
}

func (ws *Websocket) keepAlive(c *websocket.Conn, timeout time.Duration, ec chan WebsocketError) {

	switch ws.Cid.Channel {
	case Candlestick, Trade:
		return // are updated quite often
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
				ec <- WebsocketError{ws, err}
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

func (ws *Websocket) Subscribe() error {
	return Collector.Subscribe(ws)
}

func (ws *Websocket) Connect(mc chan interface{}, ec chan WebsocketError) error {
	go ws.worker(mc, ec)
	return nil
}
