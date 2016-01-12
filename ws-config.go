package ws

import "time"

var (
	/*
		Websocket options (servers & clients)
	*/
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

	// Internal
	///////////

	// Send pings to peer with this period. Must be less than PongWait.
	pingPeriod time.Duration
)
