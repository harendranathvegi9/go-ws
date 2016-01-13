package ws

import (
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Conn is a handle to underlying websocket connections.
// It allows you to send messages and close the connection,
// as well as to get info on the underlying HTTP Request.
type Conn struct {
	HTTPRequest *http.Request
	// We use an event function instead of a channel in order to
	// allow for events to be handled syncronously with the incoming
	// websocket frame stream; this ensures that the handler function
	// has access to the underlying io.Reader (which otherwise would go
	// out of scope as soon as NextReader gets called again).
	eventHandler EventHandler
	wsConn       *websocket.Conn
	sendChan     sendChan
	pingTicker   *time.Ticker
	closeOnce    sync.Once
	didClose     bool
}

// SendBinary sends a binary websocket message and returns an error if the
// connection has been closed (ErrorSendClosedConn) or if its send buffer
// is full (ErrorSendFullBuffer).
func (c *Conn) SendBinary(data []byte) (err error) {
	return c._send(BinaryMessage, data)
}

// SendText sends a text websocket message and returns an error if the
// connection has been closed (ErrorSendClosedConn) or if its send buffer
// is full (ErrorSendFullBuffer).
func (c *Conn) SendText(text string) (err error) {
	return c._send(TextMessage, []byte(text))
}
func (c *Conn) Close() {
	c._disconnect(nil)
}
func (c *Conn) String() string {
	return "{Conn " + c.HTTPRequest.URL.String() + "/" + c.HTTPRequest.RemoteAddr + "}"
}

// Internal
///////////

type sendChan chan outboundFrame
type outboundFrame struct {
	Type EventType
	data []byte
}

func init() {
	// Sanity check
	if websocket.TextMessage != EventType(TextMessage) {
		panic("Enum value mismatch: TextMessage")
	}
	if websocket.BinaryMessage != EventType(BinaryMessage) {
		panic("Enum value mismatch: BinaryMessage")
	}
}

func newConn(httpRequest *http.Request, wsConn *websocket.Conn, eventHandler EventHandler) *Conn {
	conn := &Conn{
		HTTPRequest:  httpRequest,
		wsConn:       wsConn,
		sendChan:     make(sendChan, ConnMaxSendBufferLen),
		eventHandler: eventHandler,
		pingTicker:   time.NewTicker(pingPeriod),
		closeOnce:    sync.Once{},
	}
	go _generateEvent(eventHandler, Connected, conn, nil, nil)
	go conn._writeLoop()
	go conn._readLoop()
	return conn
}

func (c *Conn) _send(msgType EventType, data []byte) error {
	if c.didClose {
		return ErrorSendClosedConn
	}
	select {
	case c.sendChan <- outboundFrame{msgType, data}:
		return nil
	default:
		return ErrorSendFullBuffer
	}
}

func (c *Conn) _writeLoop() {
	c._write(websocket.PingMessage, []byte{})
	for {
		select {
		case message, ok := <-c.sendChan:
			if !ok {
				c._disconnect(errors.New("Error reading from sendChan"))
				return
			}
			c._write(int(message.Type), message.data)
		case <-c.pingTicker.C:
			c._write(websocket.PingMessage, nil)
		}
	}
}
func (c *Conn) _write(frameType int, payload []byte) {
	err := c.wsConn.SetWriteDeadline(time.Now().Add(WriteWait))
	if err != nil {
		c._disconnect(err)
		return
	}
	err = c.wsConn.WriteMessage(frameType, payload)
	if err != nil {
		c._disconnect(err)
		return
	}
}

func (c *Conn) _readLoop() {
	// c.wsConn.SetReadLimit(512) // Maximum message size allowed from peer.
	c.wsConn.SetPongHandler(func(string) error {
		// TODO: Disconnect if err?
		return c.wsConn.SetReadDeadline(time.Now().Add(PongWait))
	})

	for {
		frameType, reader, err := c.wsConn.NextReader()
		if err != nil {
			if _, ok := (err.(*websocket.CloseError)); ok {
				// This can happen when the client websocket closes.
				// Is it a gorilla/websocket bug?
				c._disconnect(nil)

			} else if err == io.EOF {
				// Just disconnect on clean ends
				c._disconnect(nil)

				// } else if err == "connection reset by peer" {
				// } else if neterr, ok := err.(net.Error); ok {
				// 	neterr.
			} else {
				c._disconnect(err)
			}
			break
		}

		if frameType == websocket.TextMessage {
			_generateEvent(c.eventHandler, TextMessage, c, reader, nil)
		} else if frameType == websocket.BinaryMessage {
			_generateEvent(c.eventHandler, BinaryMessage, c, reader, nil)
		} else {
			_generateEvent(c.eventHandler, Error, c, nil, errors.New("Bad message type"))
		}
	}
}

func _generateEvent(eventHandler EventHandler, eventType EventType, conn *Conn, reader io.Reader, err error) {
	if eventType == Error || eventType == NetError {
		if err == nil {
			panic("Expected an error")
		}
	} else if eventType == TextMessage || eventType == BinaryMessage {
		if reader == nil || conn == nil || err != nil {
			panic("Expected a reader, a connection, and no error")
		}
	} else if eventType == Connected || eventType == Disconnected {
		if conn == nil || err != nil {
			panic("Expected a connection, and no error")
		}
	} else {
		panic("Bad event type")
	}
	event := &Event{eventType, err, reader}
	eventHandler(event, conn)
	event.reader = nil // See Event.Read
}

func (c *Conn) _disconnect(err error) {
	c.closeOnce.Do(func() {
		eventHandler := c.eventHandler
		c.eventHandler = nil
		c.pingTicker.Stop()
		c.wsConn.Close()
		go func() {
			if err != nil {
				if netError, ok := err.(net.Error); ok {
					_generateEvent(eventHandler, NetError, c, nil, netError)
				} else {
					_generateEvent(eventHandler, NetError, c, nil, err)
				}
			}
			_generateEvent(eventHandler, Disconnected, c, nil, nil)
		}()
	})
}
