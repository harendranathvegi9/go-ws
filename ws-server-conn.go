package ws

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func init() {
	if websocket.TextMessage != EventType(TextMessage) {
		panic("Enum value mismatch: TextMessage")
	}
	if websocket.BinaryMessage != EventType(BinaryMessage) {
		panic("Enum value mismatch: BinaryMessage")
	}
}

type EventHandler func(event *Event, connection *Connection)

func upgradeWebsocket(eventHandler EventHandler, w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		_generateEvent(eventHandler, Error, nil, nil, err)
		return
	}
	c := &Connection{
		Request:      r,
		wsConn:       wsConn,
		sendChan:     make(chan outgoingMessage, 256),
		eventHandler: eventHandler,
		pingTicker:   time.NewTicker(pingPeriod),
		isClosed:     false,
		mutex:        &sync.Mutex{},
	}
	go c._writeLoop()
	go c._readLoop()
	_generateEvent(eventHandler, Connected, c, nil, nil)
}

type Connection struct {
	*http.Request
	wsConn       *websocket.Conn
	sendChan     chan outgoingMessage
	eventHandler EventHandler
	pingTicker   *time.Ticker
	isClosed     bool
	mutex        *sync.Mutex
}

type outgoingMessage struct {
	messageType EventType
	data        []byte
}

func (c *Connection) SendBinary(data []byte) {
	c.sendChan <- outgoingMessage{BinaryMessage, data}
}
func (c *Connection) SendText(text string) {
	c.sendChan <- outgoingMessage{TextMessage, []byte(text)}
}
func (c *Connection) Close() {
	c._disconnect(nil)
}

func (c *Connection) String() string {
	return "{Connection " + c.Request.RemoteAddr + "}"
}

// Internal
///////////

func (c *Connection) _writeLoop() {
	c._write(websocket.PingMessage, []byte{})
	for {
		select {
		case message, ok := <-c.sendChan:
			if !ok {
				c._disconnect(errors.New("Error reading from sendChan"))
				return
			}
			c._write(message.messageType, message.data)
		case <-c.pingTicker.C:
			c._write(websocket.PingMessage, []byte{})
		}
	}
}
func (c *Connection) _write(messageType EventType, payload []byte) {
	err := c.wsConn.SetWriteDeadline(time.Now().Add(WriteWait))
	if err != nil {
		c._disconnect(err)
		return
	}
	err = c.wsConn.WriteMessage(int(messageType), payload)
	if err != nil {
		c._disconnect(err)
		return
	}
}

func (c *Connection) _readLoop() {
	// c.wsConn.SetReadLimit(512) // Maximum message size allowed from peer.
	c.wsConn.SetPongHandler(func(string) error {
		// TODO: Disconnect if err?
		return c.wsConn.SetReadDeadline(time.Now().Add(PongWait))
	})

	for {
		messageType, reader, err := c.wsConn.NextReader()
		if err != nil {
			if _, ok := (err.(*websocket.CloseError)); ok {
				// This can happen when the client websocket closes.
				// Is it a gorilla/websocket bug?
				c._disconnect(nil)

			} else if err == io.EOF {
				// Just disconnect on clean ends
				c._disconnect(nil)

			} else {
				c._disconnect(err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			_generateEvent(c.eventHandler, TextMessage, c, reader, nil)
		} else if messageType == websocket.BinaryMessage {
			_generateEvent(c.eventHandler, BinaryMessage, c, reader, nil)
		} else {
			_generateEvent(c.eventHandler, Error, c, nil, errors.New("Bad message type"))
		}
	}
}

func _generateEvent(eventHandler EventHandler, eventType EventType, conn *Connection, reader io.Reader, err error) {
	if eventType == Error {
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

func (c *Connection) _disconnect(err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.isClosed {
		return
	}
	c.isClosed = true
	eventHandler := c.eventHandler
	c.eventHandler = nil
	c.wsConn.Close()
	c.pingTicker.Stop()
	close(c.sendChan)
	if err != nil {
		_generateEvent(eventHandler, Error, c, nil, err)
	}
	_generateEvent(eventHandler, Disconnected, c, nil, nil)
}
