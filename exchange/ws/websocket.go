package ws

import (
	"fmt"
	"github.com/google/logger"
	"github.com/gorilla/websocket"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	wsTimeout = time.Second * 60
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
	conn *websocket.Conn
	handler Handler
	cClose chan struct{}
	m *sync.Mutex
	pp bool
}

func Connect(handler Handler) {

	//
	// connection timout workaround
	// required because Dial can stall too long
	//    on dns resolving / tcp connecting
	//

	if atomic.LoadInt32(&started) == 0 {
		go ceWorker()
	}

	wsc := make(chan struct{*websocket.Conn;error})

	go func() {
		var c *websocket.Conn
		var err error
		defer func() {
			if r := recover(); r != nil {
				if err == nil && c != nil {
					logger.Infof("to late: %#v",c)
					_ = c.Close()
				}else{
					logger.Infof("to late: %v",err.Error())
				}
			}
		}()
		logger.Info("connecting")
		c, _, err = websocket.DefaultDialer.Dial(handler.Endpoint(), nil)
		wsc <- struct{*websocket.Conn;error}{c,err}
		close(wsc)
	}()

	connected := func(q struct{*websocket.Conn;error}) {
		if q.error != nil {
			fatal := false
			if strings.Index(q.error.Error(), "malformed") >= 0 {
				fatal = true
			}
			ce <- &conError{handler, q.error, fatal }
		} else {
			ws := &Websocket{q.Conn, handler, make(chan struct{}), &sync.Mutex{}, false }
			fatal,err := handler.OnConnect(ws)
			ce <- &conError{handler, err, fatal }
			go ws.worker()
		}
	}

	for {
		select {
		case <-time.After(3 * time.Second):
			logger.Infof("connection timeout")
			close(wsc)
			if q, ok:= <-wsc; ok {
				connected(q)
				return
			}
			ce <- &conError { handler, fmt.Errorf("connection timeout"), false }
			return
		case q := <-wsc:
			connected(q)
			return
		}
	}
}

func (ws *Websocket) KeepAlive() {
	ws.pp = true
}

func (ws *Websocket) Close() error {
	ws.m.Lock()
	if ws.cClose != nil {
		ws.cClose <- struct{}{}
		ws.cClose = nil
	}
	ws.m.Unlock()
	return nil
}

func (ws *Websocket) worker() {
	var isBroken int32
	ticker := time.NewTicker(wsTimeout)
	c := ws.cClose

	defer func() {
		ticker.Stop()
		ws.m.Lock(); ws.cClose = nil; ws.m.Unlock();
		close(c);
		_ = ws.conn.Close()
		ws.handler.OnDisconnect()
	}()

	lastResponse := time.Now()
	ws.conn.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	if ws.pp {
		go func() {
			for atomic.LoadInt32(&isBroken) == 0 {
				select {
				case <-c:
					return
				case <-ticker.C:
					if time.Now().Sub(lastResponse) > wsTimeout {
						if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
							ce <- &conError{ws.handler, fmt.Errorf("ping/pong timeout"), false}
						}
						return
					}
					deadline := time.Now().Add(30 * time.Second)
					err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, deadline)
					if err != nil {
						if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
							ce <- &conError{ws.handler, err, false}
						}
						return
					}
				}
			}
		}()
	}

	for atomic.LoadInt32(&isBroken) == 0 {
		_, message, err := ws.conn.ReadMessage();
		select {
		case <-c:
			return
		default:
			// nothing
		}
		if err != nil {
			if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
				ce <- &conError { ws.handler, fmt.Errorf("connection timeout"), false }
			}
			return
		} else if !ws.handler.OnMessage(message) {
			//if atomic.CompareAndSwapInt32(&isBroken, 0, 1) {
			//	ce <- &conError{ws.handler, fmt.Errorf("message error"), false}
			//}
			//return
		}
	}
}

