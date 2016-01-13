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
	// A text message was received.
	// Use Event.Text() to read the string.
	TextMessage = EventType(1)

	// A binary frame was received.
	// Use Event.Data() to read the data.
	BinaryMessage = EventType(2)

	// The websocket connected.
	// You are online!
	Connected = EventType(3)

	// The websocket disconnected. It should no longer be used.
	Disconnected = EventType(4)

	// A network error occured.
	// This will always be followed by a Disconnected event.
	// These events can generally be ignored.
	NetError = EventType(5)

	// An unforeseen error occured.
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

// Internal
///////////

var eventTypeNames = map[EventType]string{
	Connected:     "<Connected>",
	Disconnected:  "<Disconnected>",
	TextMessage:   "<TextMessage>",
	BinaryMessage: "<BinaryMessage>",
	Error:         "<Error>",
	NetError:      "<NetError>",
}

var eventHasReader = map[EventType]bool{
	TextMessage:   true,
	BinaryMessage: true,
}

func (e EventType) String() string { return eventTypeNames[e] }
