package ws_test

import (
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/marcuswestin/go-ws"
)

var addr = "localhost:8087"
var dialer = websocket.Dialer{}
var serverURL = "http://" + addr + "/ws/echo"

func TestListenAndServe(t *testing.T) {
	ws.UpgradeRequests("/ws/echo", func(event *ws.Event, conn ws.Conn) {
		switch event.Type {
		case ws.Connected:
			log.Println("Server received connection:", conn)

		case ws.TextMessage:
			text, err := event.Text()
			log.Println("Text message:", text, err)
			if text == "Disconnect me" {
				conn.Close()
			} else {
				conn.SendText(text)
			}

		case ws.BinaryMessage:
			data, err := event.Data() // Or use `event` as an `io.Reader`
			log.Println("Binary message size:", len(data), err)
			conn.SendBinary(data)

		case ws.Error:
			panic("Server error Conn: " + conn.HTTPRequest().RemoteAddr + ", Error: " + event.Error.Error())

		case ws.NetError:
			log.Println("Server saw net error:", conn, event.Error)

		case ws.Disconnected:
			log.Println("Server saw disconnect:", conn)
		}
	})
	go http.ListenAndServe(addr, nil)
}

type EventChan chan TestEvent
type TestEvent struct {
	Type ws.EventType
	Data []byte
}

func connect(t *testing.T) (testConn ws.Conn, eventChan EventChan) {
	eventChan = make(EventChan)
	ws.Connect(serverURL, func(event *ws.Event, conn ws.Conn) {
		log.Println("Client: Event", event.Type)
		assert(t, event.Type != ws.Error, "Received error event", event.Error)
		testConn = conn
		data, err := event.Data()
		assert(t, err == nil, "Unable to read event data", err)
		eventChan <- TestEvent{event.Type, data}
	})
	receive(t, eventChan, ws.Connected)
	log.Println("Client connected:", testConn)
	return
}

func TestServerAndClient(t *testing.T) {
	conn, eventChan := connect(t)
	sendRecv(t, conn, eventChan, ws.BinaryMessage)
	sendRecv(t, conn, eventChan, ws.TextMessage)
	conn.SendText("Disconnect me")
	receive(t, eventChan, ws.Disconnected)
}

func TestClientDisconnect(t *testing.T) {
	conn, eventChan := connect(t)
	conn.Close()
	receive(t, eventChan, ws.Disconnected)
	time.Sleep(150 * time.Millisecond)
}

func TestParallelClients(t *testing.T) {
	syncTest(t, 10)
	syncTest(t, 100)
}

func syncTest(t *testing.T, parallelCount int) {
	wg := sync.WaitGroup{}
	for i := 0; i < parallelCount; i++ {
		wg.Add(1)
		go func() {
			TestServerAndClient(t)
			wg.Done()
		}()
	}
	wg.Wait()
}

// Util
///////

func assert(t *testing.T, ok bool, msg ...interface{}) {
	if !ok {
		// t.Fatal("assert failed", msg)
		log.Panic(msg...)
	}
}

func receive(t *testing.T, eventChan EventChan, eventType ws.EventType) TestEvent {
	log.Println("Client: wait for", eventType)
	event := <-eventChan
	assert(t, event.Type == eventType, "Bad event type. Expected:", eventType, "Received:", event.Type)
	log.Println("Client: received", eventType)
	return event
}

func sendRecv(t *testing.T, conn ws.Conn, eventChan EventChan, messageType ws.EventType) {
	const message = "Hello World!"
	if messageType == ws.BinaryMessage {
		conn.SendBinary([]byte(message))
	} else {
		conn.SendText(message)
	}
	event := receive(t, eventChan, messageType)
	assert(t, string(event.Data) == message)
}
