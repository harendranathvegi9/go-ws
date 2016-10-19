package ws

import (
	"errors"
	"time"
)

var (
	/*
		Websocket options (servers & clients)
	*/

	// WriteWait is the time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// PongWait is the time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// MaxMessageSize is the maximum message size allowed from peer.
	MaxMessageSize = int64(512)

	// ReadBufferSize is the size of websocket upgrader
	ReadBufferSize = 4096

	// WriteBufferSize is the size of websocket upgrader
	WriteBufferSize = 4096

	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout = time.Duration(0)

	// ConnMaxSendBufferLen is the Maximum number of messages to buffer per
	// connection before SendText and SendBinary starts dropping messages and
	// return ErrorSendFullBuffer.
	ConnMaxSendBufferLen = 64
)

var (
	/*
		Errors
	*/

	// ErrorSendClosedConn is returned by SendText and SendBinary if called
	// on a closed connection.
	ErrorSendClosedConn = errors.New("send on closed connection")

	// ErrorSendFullBuffer is returned by SendText and SendBinary if called
	// on a connection with a full buffer.
	ErrorSendFullBuffer = errors.New("send on connection with full buffer")
)
