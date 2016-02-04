package ws

import (
	"errors"
	"io"
	"io/ioutil"
)

// EventHandler is a function that gets called once for each connection event.
// Each connection's EventHandler function is guaranteed to be called serially
// per-connection, but may be called concurrently across multiple connections.
type EventHandler func(event *Event, conn *Conn)

// EventType is the type of websocket event:
// Connected, TextMessage, BinaryMessage, Error, NetError, Disconnected
type EventType uint8

// Event encapsulates the information passed to EventHandlers.
// Note that Event.Text, Event.Data and Event.Read may only be
// called during the duriduration of its EventHandler function.
// As soon as the EventHandler returns, the underlying io.Reader
// will no longer be valid.
type Event struct {
	Type   EventType
	Error  error
	reader io.Reader
}

var (
	/*
		Event types.

		These signifiy the different types of events that
		happen during the lifetime of an underlying websocket.
	*/

	// TextMessage signifies that a text message was received.
	// Use Event.Text() to read the string.
	TextMessage = EventType(1)

	// BinaryMessage signifies that a binary frame was received.
	// Use Event.Data() to read the data.
	BinaryMessage = EventType(2)

	// Connected signifies that the websocket connected.
	// You are online!
	Connected = EventType(3)

	// Disconnected signifies that the websocket disconnected.
	// It should no longer be used.
	Disconnected = EventType(4)

	// NetError signifies that a network error occured.
	// This will always be followed by a Disconnected event.
	// These events can generally be ignored.
	NetError = EventType(5)

	// Error signifies that an unforeseen error occured.
	// This will always be followed by a Disconnected event.
	Error = EventType(6)
)

// Text reads the data of the underlying text message
// frame and returns it as a string.
func (e *Event) Text() (string, error) {
	if e.Type != TextMessage {
		return "", errors.New("Event.Text() called on non-TextMessage event")
	}
	bts, err := ioutil.ReadAll(e)
	return string(bts), err
}

// Data reads the data of the underlying binary message
// frame and returns it.
func (e *Event) Data() ([]byte, error) {
	if e.Type != BinaryMessage {
		return nil, errors.New("Event.Text() called on non-BinaryMessage event")
	}
	return ioutil.ReadAll(e)
}

// Event implements Reader.
func (e *Event) Read(p []byte) (int, error) {
	if !eventHasReader[e.Type] {
		return 0, io.EOF
	}
	if e.reader == nil {
		return 0, errors.New("Read may only be called during EventHandler (See https://godoc.org/github.com/marcuswestin/go-ws#Event)")
	}
	return e.reader.Read(p)
}

// String returns a string representation of Event that is suitable for debugging
func (e *Event) String() string {
	if e.Type == Error {
		return "{ws.Event Error " + e.Error.Error() + "}"
	}
	return "{ws.Event " + e.Type.String() + "}"
}

// Internal
///////////

var eventTypeNames = map[EventType]string{
	Connected:     "<EventConnected>",
	Disconnected:  "<EventDisconnected>",
	TextMessage:   "<EventTextMessage>",
	BinaryMessage: "<EventBinaryMessage>",
	Error:         "<EventError>",
	NetError:      "<EventNetError>",
}

var eventHasReader = map[EventType]bool{
	TextMessage:   true,
	BinaryMessage: true,
}

func (e EventType) String() string {
	if name, exists := eventTypeNames[e]; exists {
		return name
	}
	panic("EventType.String unknown event type")
}

func _checkAndGenerateEvent(eventHandler EventHandler, eventType EventType, conn *Conn, reader io.Reader, err error) {
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
