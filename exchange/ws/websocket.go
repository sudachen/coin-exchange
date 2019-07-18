package ws

import (
	"bytes"
	"compress/flate"
	"fmt"
	"github.com/google/logger"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	wsTimeout = time.Second * 120
)

type Handler interface {
	OnConnect(ws *Websocket) (bool, error)
	OnDisconnect()
	OnMessage(msg []byte) bool
	OnFatal(err error)
	Endpoint() string
	String() string
}

type Websocket struct {
	conn    *websocket.Conn
	handler Handler
	cClose  chan struct{}
	m       *sync.Mutex
	pp      bool
}

func connect(handler Handler) {

	//
	// connection timout workaround
	// required because Dial can stall too long
	//    on dns resolving / tcp connecting
	//

	if atomic.LoadInt32(&started) == 0 {
		go ceWorker()
	}

	wsc := make(chan struct {
		*websocket.Conn
		error
	})

	go func() {
		var c *websocket.Conn
		var err error
		defer func() {
			if r := recover(); r != nil {
				if err == nil && c != nil {
					logger.Infof("to late: %#v", c)
					_ = c.Close()
				} else {
					logger.Infof("to late: %v", err.Error())
				}
			}
		}()
		c, _, err = websocket.DefaultDialer.Dial(handler.Endpoint(), nil)
		wsc <- struct {
			*websocket.Conn
			error
		}{c, err}
		close(wsc)
	}()

	connected := func(q struct {
		*websocket.Conn
		error
	}) {
		if q.error != nil {
			fatal := false
			if strings.Index(q.error.Error(), "malformed") >= 0 {
				fatal = true
			}
			ce <- &conError{handler, q.error, fatal}
		} else {
			ws := &Websocket{q.Conn, handler, make(chan struct{}), &sync.Mutex{}, false}
			fatal, err := handler.OnConnect(ws)
			ce <- &conError{handler, err, fatal}
			go ws.worker()
		}
	}

	for {
		select {
		case <-time.After(3 * time.Second):
			logger.Infof("connection timeout")
			close(wsc)
			if q, ok := <-wsc; ok {
				connected(q)
				return
			}
			ce <- &conError{handler, fmt.Errorf("connection timeout"), false}
			return
		case q := <-wsc:
			connected(q)
			return
		}
	}
}

func Connect(handler Handler) {
	go connect(handler)
}

func (ws *Websocket) Send(bs []byte) error {
	return ws.conn.WriteMessage(websocket.TextMessage, bs)
}

func (ws *Websocket) Conn() *websocket.Conn {
	return ws.conn
}

func (ws *Websocket) KeepAlive() {
	ws.pp = true
}

func (ws *Websocket) Close() error {
	ws.m.Lock()
	if ws.cClose != nil {
		close(ws.cClose)
		ws.cClose = nil
	}
	ws.m.Unlock()
	return nil
}

func (ws *Websocket) worker() {
	var isBroken int32
	ticker := time.NewTicker(wsTimeout)
	cClose := ws.cClose

	defer func() {
		ticker.Stop()
		_ = ws.Close() // close cClose on network error
		logger.Infof("stream disconnected %v", ws.handler.String())
		ws.handler.OnDisconnect()
	}()

	lastResponse := time.Now()
	ws.conn.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		// eeeh...
		// looks like I have to brake off connection here
		defer func() { _ = ws.conn.Close() } ()

		for atomic.LoadInt32(&isBroken) == 0 {
			select {
			case <-cClose:
				return
			case <-ticker.C:
				if ws.pp { // if we have to check ppp
					if time.Now().Sub(lastResponse) > wsTimeout {
						if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
							ce <- &conError{ws.handler, fmt.Errorf("ping/pong timeout"), false}
						}
						return
					}
					deadline := time.Now().Add(10 * time.Second)
					err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, deadline)
					if err != nil {
						if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
							ce <- &conError{ws.handler, err, false}
						}
						return
					}
				}
			}
		}
	}()

	for atomic.LoadInt32(&isBroken) == 0 {
		t, message, err := ws.conn.ReadMessage()
		select {
		case <-cClose:
			return
		default:
			// nothing
		}
		if err != nil {
			if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
				ce <- &conError{ws.handler, fmt.Errorf("connection timeout"), false}
			}
			return
		} else {
			if t == websocket.BinaryMessage {
				r := flate.NewReader(bytes.NewReader(message))
				message, err = ioutil.ReadAll(r)
				_ = r.Close()
			}
			_ = ws.handler.OnMessage(message)
		}
	}
}
