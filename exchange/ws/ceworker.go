package ws

import (
	"fmt"
	"github.com/google/logger"
	"sync/atomic"
	"time"
)

const (
	maxReconnectCount = 10
	slowReconnectCount = 5
)

const (
	reconnectTickerTimeout = time.Second*5
	fastReconnectTimeout = time.Second*10
	slowReconnectTimeout = time.Minute*5
)

const ErrQueueLength = 13

type conError struct {
	h Handler
	err error
	fatal bool
}

var started int32
var ce = make(chan *conError, ErrQueueLength)

func ceWorker() {

	if !atomic.CompareAndSwapInt32(&started, 0, 1) {
		return
	}

	ticker := time.NewTicker(reconnectTickerTimeout)
	defer ticker.Stop()

	type RS struct {
		till time.Time
		waitForReconnect bool
		reconnectCount int
	}

	hndls := make(map[Handler]RS)

	logger.Info("websocket error handler started")

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for h, rs := range hndls {
				if rs.waitForReconnect && rs.till.Sub(now) <= 0 {
					rs.waitForReconnect = false
					hndls[h] = rs
					Connect(h)
				}
			}

		case e := <-ce:
			rs, exists := hndls[e.h]
			if e.err == nil {
				logger.Infof("stream connected successful %v", e.h.String())
				if exists {
					delete(hndls,e.h)
				}
			} else if e.fatal {
				logger.Errorf("received fatal error from %v: %v",
					e.h.String(),
					e.err.Error())
				e.h.OnFatal(e.err)
				delete(hndls,e.h)
			} else {
				logger.Warningf("received non-fatal error from %v: %v",
					e.h.String(),
					e.err.Error())
				if rs.waitForReconnect && rs.reconnectCount == maxReconnectCount {
					logger.Errorf("reconnection limit exceeded %v", e.h.String())
					e.h.OnFatal(fmt.Errorf("reconnection limit exceeded"))
				} else {
					if !exists {
						rs = RS{}
					}
					logger.Infof("will reconnect %v", e.h.String())
					logger.Infof("%#v", rs)
					rs.reconnectCount += 1
					if rs.reconnectCount < 2 {
						rs.waitForReconnect = false
						logger.Infof("reconnecting %v", e.h.String())
						Connect(e.h)
					} else {
						rs.waitForReconnect = true
						if rs.reconnectCount < slowReconnectCount {
							rs.till = time.Now().Add(fastReconnectTimeout)
						} else {
							rs.till = time.Now().Add(slowReconnectTimeout)
						}
					}
				}
				hndls[e.h] = rs
			}
		}
	}
}
