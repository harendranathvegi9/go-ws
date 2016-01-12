package ws

import (
	"strings"

	"github.com/gorilla/websocket"
)

func Connect(addr string, eventHandler EventHandler) {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		addr = strings.Replace(addr, "http", "ws", 1)
	}
	wsConn, httpRes, err := dialer.Dial(addr, nil)
	if err != nil {
		_generateEvent(eventHandler, Error, nil, nil, err)
		return
	}
	newConn(httpRes.Request, wsConn, eventHandler) // sets up read/write loop
}

// Internal
///////////

var dialer = websocket.Dialer{}
