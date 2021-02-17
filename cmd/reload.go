package cmd

import (
	"log"
	"plenti/common"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
)

var reloadC = make(chan struct{}, 1)

// keep all connections in slice for broadcast
var connections = map[*websocket.Conn]bool{}

// lock for ^^
var connMU sync.Mutex

// count of loaded to come back
var numReloading int32

// wshandler handles the numReloading when serve -L is used
func wshandler(ws *websocket.Conn) {
	// add new conn
	connMU.Lock()
	connections[ws] = true
	connMU.Unlock()
	// catches when browser/tab is closed.
	// ws.Request.Context.Done() is always nil so have ping
	// might do more frequent?
	pinger := time.NewTicker(time.Second * 5)

	defer func() {

		pinger.Stop()
		ws.Close()
	}()

	go func() {
		var msg string

		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Println(err, "error in receive")

		}

		connMU.Lock()
		// atomic.AddInt32 not really needed as we lock... do need the Lock() for if numReloading...
		if numReloading > 0 && atomic.AddInt32(&numReloading, -1) == 0 {
			// page(s) is loaded so unlock and allow next build
			common.Unlock()

		}
		connMU.Unlock()

	}()
	for {
		select {
		case <-pinger.C:
			// presume no news is good news. err/news == closed i.e refresh/tab/browser closed
			if err := websocket.Message.Send(ws, "ping"); err != nil {
				// once less for reloading
				atomic.AddInt32(&numReloading, -1)
				connMU.Lock()
				// remove if not already gone. no-op if already emptied
				delete(connections, ws)
				connMU.Unlock()

				return

			}

		case <-reloadC:
			connMU.Lock()
			// reset
			numReloading = 0
			// broadcast to all
			for wlst := range connections {
				err := websocket.Message.Send(wlst, "reload")
				// presume closed/disconnected on error so don't inc numReloading
				// will be a tie between pings that could appear in connections but actually gone
				// todo: check error and log  somewhere
				if err != nil {
					continue

				}
				atomic.AddInt32(&numReloading, 1)

			}
			/// empty all conns
			for k := range connections {
				delete(connections, k)
			}
			connMU.Unlock()
			// close as new connection each reload, otherwise broken pipe errors etc..
			return

		}
	}

}
