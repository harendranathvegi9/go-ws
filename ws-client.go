package ws

import (
	"strings"

	"github.com/gorilla/websocket"
	"github.com/marcuswestin/go-errs"
)

// Connect opens a websocket connection to the given address,
// and starts generating events.
func Connect(addr string, eventHandler EventHandler) {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		addr = strings.Replace(addr, "http", "ws", 1)
	}
	wsConn, httpRes, err := dialer.Dial(addr, nil)
	if err != nil {
		_checkAndGenerateEvent(eventHandler, Error, nil, nil, err)
		return
	}
	if wsConn.Subprotocol() != "birect" {
		_checkAndGenerateEvent(eventHandler, Error, nil, nil, errs.New(nil, "Unable to negotiate birect subprotocol"))
		return
	}
	newConn(httpRes.Request, wsConn, eventHandler) // sets up read/write loop
}

// Internal
///////////

var dialer = websocket.Dialer{
	Subprotocols: []string{"birect"},
}
