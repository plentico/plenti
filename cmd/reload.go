package cmd

import (
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

var reloadC = make(chan struct{}, 1)

type wsc struct {
	ws     *websocket.Conn
	closeC chan struct{}
}

// keep all connections in map for broadcast
var connections = map[wsc]bool{}

// lock for ^^
var connMU sync.Mutex

// count of loaded to come back
var numReloading int32

// wshandler handles the numReloading when serve -L is used
func wshandler(ws *websocket.Conn) {
	// add new conn and chan to close
	cc := wsc{ws: ws, closeC: make(chan struct{}, 1)}
	connMU.Lock()
	// inc new conn
	atomic.AddInt32(&numReloading, 1)
	connections[cc] = true
	connMU.Unlock()

	pinger := time.NewTicker(time.Second * 5)
	defer func() {

		pinger.Stop()
		ws.Close()

	}()

	// just read once...
	go func() {
		var msg string
		ws.SetReadDeadline(time.Now().Add(time.Second * 5))
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Println(err, "error in receive")
		}
	}()
	for {
		select {
		// close and decrement count
		case <-cc.closeC:
			cc.closeC = nil
			return
		case <-pinger.C:

			ws.SetWriteDeadline(time.Now().Add(time.Second * 10))
			// presume no news is good news. err/news == closed i.e refresh/tab/browser closed
			if err := websocket.Message.Send(ws, "ping"); err != nil {

				connMU.Lock()

				delete(connections, cc)

				connMU.Unlock()

				return

			}

		case <-reloadC:

			connMU.Lock()

			// broadcast to all
			for wlst := range connections {
				// not sure how important this is for way we use
				wlst.ws.SetWriteDeadline(time.Now().Add(time.Second * 10))
				err := websocket.Message.Send(wlst.ws, "reload")
				// todo:  check error and log somewhere
				_ = err

			}

			// empty all conns and close
			for k := range connections {
				// close each conn individually. Should catch edge cases where closing tabs/browsers even while reloading..
				k.closeC <- struct{}{}
				delete(connections, k)
			}

			connMU.Unlock()
			return

		}

	}

}

// debugging
func paddr(ws *websocket.Conn) string {
	return ws.Request().RemoteAddr[strings.LastIndex(ws.Request().RemoteAddr, ":")+1:]
}
