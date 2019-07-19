package ws

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
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
	wsTimeout = 60
)

type Handler interface {
	OnConnect(ws *Websocket) (bool, error)
	OnDisconnect()
	OnMessage(msg []byte) bool
	OnFatal(err error)
	Endpoint() string
	String() string
}

type Compression byte

const (
	Defalted Compression = iota
	Gzipped
)

type Websocket struct {
	conn    *websocket.Conn
	handler Handler
	cClose  chan struct{}
	m       *sync.Mutex
	pp      func(*Websocket) error
	out     chan []byte
	Compression
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
			ws := &Websocket{
				q.Conn,
				handler,
				make(chan struct{}),
				&sync.Mutex{},
				nil,
				make(chan []byte, 3),
				Defalted}

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
	//logger.Infof("WS<=%s\n",string(bs))
	ws.out <- bs
	return nil
}

func (ws *Websocket) Conn() *websocket.Conn {
	return ws.conn
}

func (ws *Websocket) KeepAlive(kaf func(*Websocket) error) {
	ws.pp = kaf
}

func (ws *Websocket) Ping() error {
	deadline := time.Now().Add(10 * time.Second)
	ws.m.Lock()
	err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, deadline)
	ws.m.Unlock()
	return err
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
	ticker := time.NewTicker(wsTimeout * time.Second / 2)
	cClose := ws.cClose

	defer func() {
		ticker.Stop()
		_ = ws.Close() // close cClose on network error
		logger.Infof("stream disconnected %v", ws.handler.String())
		ws.handler.OnDisconnect()
	}()

	var lastResponse int64 = time.Now().Unix()
	setResponseTime := func() {
		i := time.Now().Unix()
		atomic.StoreInt64(&lastResponse, i)
	}

	ws.conn.SetPongHandler(func(msg string) error {
		setResponseTime()
		return nil
	})

	go func() {
		// eeeh...
		// looks like I have to brake off connection here
		defer func() { _ = ws.conn.Close() }()

		for atomic.LoadInt32(&isBroken) == 0 {
			select {
			case <-cClose:
				return
			case m := <-ws.out:
				err := ws.conn.WriteMessage(websocket.TextMessage, m)
				if err != nil {
					if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
						ce <- &conError{ws.handler, err, false}
					}
					return
				}
			case <-ticker.C:
				if ws.pp != nil { // if we have to check ppp
					if (time.Now().Unix() - atomic.LoadInt64(&lastResponse)) > wsTimeout {
						if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
							ce <- &conError{ws.handler, fmt.Errorf("keepalive timeout"), false}
						}
						return
					}
					err := ws.pp(ws)
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
		setResponseTime()
		if err != nil {
			if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
				ce <- &conError{ws.handler, fmt.Errorf("network error: %v", err), false}
			}
			return
		} else {
			if t == websocket.BinaryMessage {
				if ws.Compression == Gzipped {
					r, _ := gzip.NewReader(bytes.NewBuffer(message))
					message, err = ioutil.ReadAll(r)
					_ = r.Close()
				} else {
					r := flate.NewReader(bytes.NewReader(message))
					message, err = ioutil.ReadAll(r)
					_ = r.Close()
				}
			}
			_ = ws.handler.OnMessage(message)
		}
	}
}
