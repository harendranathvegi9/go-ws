package ws

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// Maximum message size allowed from peer.
	MaxMessageSize = int64(512)

	// Read buffer size of websocket upgrader
	ReadBufferSize = 4096

	// Write buffer size of websocket upgrader
	WriteBufferSize = 4096

	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout = time.Duration(0)

	// Subprotocols specifies the server's supported protocols in order of
	// preference. If this option is set, then the upgrade handler negotiates a
	// subprotocol by selecting the first match in this list with a protocol
	// requested by the client.
	Subprotocols []string

	// ErrorFn specifies the function for generating HTTP error responses. If ErrorFn
	// is nil, then http.Error is used to generate the HTTP response.
	ErrorFn func(w http.ResponseWriter, r *http.Request, status int, reason error)

	CheckOrigin = func(r *http.Request) bool { return true }
)

// Send pings to peer with this period. Must be less than PongWait.
var pingPeriod time.Duration
var upgrader websocket.Upgrader
var called = false // Hack

func UpgradeRequests(pattern string, eventHandler EventHandler) {
	if called {
		panic("UpgradeRequests should be called once")
	}
	called = true
	pingPeriod = (PongWait * 7) / 10
	upgrader = websocket.Upgrader{
		HandshakeTimeout,
		ReadBufferSize,
		WriteBufferSize,
		Subprotocols,
		ErrorFn,
		CheckOrigin,
	}
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		upgradeWebsocket(eventHandler, w, r)
	})
	return
}

func upgradeWebsocket(eventHandler EventHandler, w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		_generateEvent(eventHandler, Error, nil, nil, err)
		return
	}
	newConn(r, wsConn, eventHandler)
}
