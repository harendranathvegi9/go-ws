package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/marcuswestin/go-errs"
)

var (
	/*
		HTTP Upgrade Options

		These are the different options
	*/

	// CheckOrigin will be called by the upgrader on incoming requests to check
	// the origin. If the CheckOrigin function returns false, then the Upgrade
	// method fails the WebSocket handshake with HTTP status 403.
	// If nil, then use a safe default: fail the handshake if the Origin request
	// header is present and not equal to the Host request header.
	CheckOrigin func(r *http.Request) bool

	// ErrorFn specifies the function for generating HTTP error responses.
	// If nil, then http.Error is used to generate the HTTP response.
	ErrorFn func(w http.ResponseWriter, r *http.Request, status int, reason error)
)

// UpgradeRequests will upgrade any incoming websocket request
// matching the given pattern, and then call the event handler
// function with connection events.
func UpgradeRequests(pattern string, eventHandler EventHandler) {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: HandshakeTimeout,
		ReadBufferSize:   ReadBufferSize,
		WriteBufferSize:  WriteBufferSize,
		Subprotocols:     []string{"birect"},
		Error:            ErrorFn,
		CheckOrigin:      CheckOrigin,
	}
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if version := r.Header.Get("Sec-Websocket-Version"); version != "13" {
			http.Error(w, "Sec-Websocket-Version must be 13", 400)
			return
		}
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			_checkAndGenerateEvent(eventHandler, Error, nil, nil, err)
		} else if wsConn.Subprotocol() != "birect" {
			_checkAndGenerateEvent(eventHandler, Error, nil, nil, errs.New(nil, "Unable to negotiate birect subprotocol"))
		} else {
			newConn(r, wsConn, eventHandler)
		}
	})
	return
}
