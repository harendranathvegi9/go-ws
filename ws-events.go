package ws

import (
	"errors"
	"io"
	"io/ioutil"
)

type EventChan chan Event
type EventType uint8
type Event struct {
	Type   EventType
	Error  error
	reader io.Reader
}

var (
	TextMessage   = EventType(1)
	BinaryMessage = EventType(2)
	Connected     = EventType(3)
	Disconnected  = EventType(4)
	Error         = EventType(5)
	NetError      = EventType(6)
)

func (e *Event) Text() (string, error) {
	bts, err := e.Data()
	return string(bts), err
}
func (e *Event) Data() ([]byte, error) {
	return ioutil.ReadAll(e)
}

// Event implements Reader. However, it can only be read during
// the duration of the EventHandler function in which Event was
// given. As soon as the EventHandler returns, the io.Reader will no
// longer be valid.
func (e *Event) Read(p []byte) (int, error) {
	if !eventHasReader[e.Type] {
		return 0, io.EOF
	}
	if e.reader == nil {
		return 0, errors.New("Read may only be called during EventHandler (See https://godoc.org/github.com/marcuswestin/go-ws#Event.Read)")
	}
	return e.reader.Read(p)
}

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
