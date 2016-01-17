package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	/*
		HTTP Upgrade Options

		These are the different options
	*/
	// The Upgrader calls the function specified in the CheckOrigin field to check
	// the origin. If the CheckOrigin function returns false, then the Upgrade
	// method fails the WebSocket handshake with HTTP status 403.
	// If nil, then use a safe default: fail the handshake if the Origin request
	// header is present and not equal to the Host request header.
	CheckOrigin func(r *http.Request) bool

	// ErrorFn specifies the function for generating HTTP error responses.
	// If nil, then http.Error is used to generate the HTTP response.
	ErrorFn func(w http.ResponseWriter, r *http.Request, status int, reason error)

	// Subprotocols specifies the server's supported protocols in order of
	// preference.
	// If set, the upgrade handler negotiates a subprotocol by selecting the
	// first match in this list with a protocol requested by the client.
	Subprotocols []string
)

// UpgradeRequests will upgrade any incoming websocket request
// matching the given pattern, and then call the event handler
// function with connection events.
func UpgradeRequests(pattern string, eventHandler EventHandler) {
	pingPeriod = (PongWait * 7) / 10
	upgrader := websocket.Upgrader{
		HandshakeTimeout,
		ReadBufferSize,
		WriteBufferSize,
		Subprotocols,
		ErrorFn,
		CheckOrigin,
	}
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			_checkAndGenerateEvent(eventHandler, Error, nil, nil, err)
		} else {
			newConn(r, wsConn, eventHandler)
		}
	})
	return
}
