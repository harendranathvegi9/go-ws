package ws_test

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/marcuswestin/go-ws"
)

var addr = "localhost:8087"
var dialer = websocket.Dialer{}

func TestListenAndServe(t *testing.T) {
	ws.UpgradeRequests("/ws/echo", func(event *ws.Event, connection *ws.Connection) {
		switch event.Type {
		case ws.TextMessage, ws.BinaryMessage:
			bts, err := event.Bytes()
			assert(t, err == nil)
			connection.Send(event.Type, bts) // Echo

		case ws.Error:
			panic("Error:" + event.Error.Error())

		case ws.Connected:
			log.Println("Server: Connected")

		case ws.Disconnected:
			log.Println("Server: Disconnected")
		}
	})
	go http.ListenAndServe(addr, nil)
}

func TestServerAndClient(t *testing.T) {
	ws, _, err := dialer.Dial("ws://"+addr+"/ws/echo", nil)
	assert(t, err == nil, err)
	sendRecv(t, ws)
	ws.Close()
	time.Sleep(50 * time.Millisecond) // give server time to handle close
}

// Util
///////

func assert(t *testing.T, ok bool, msg ...interface{}) {
	if !ok {
		t.Fatal("assert failed", msg)
		log.Panic(msg...)
	}
}

func sendRecv(t *testing.T, ws *websocket.Conn) {
	const message = "Hello World!"
	err := ws.SetWriteDeadline(time.Now().Add(time.Second))
	assert(t, err == nil, err)

	log.Println("Client: Send message", message)
	err = ws.WriteMessage(websocket.BinaryMessage, []byte(message))
	assert(t, err == nil, err)
	err = ws.SetReadDeadline(time.Now().Add(time.Second))
	assert(t, err == nil, err)

	log.Println("Client: Read next message")
	messageType, bts, err := ws.ReadMessage()
	log.Println("Client: Received message", message)
	assert(t, err == nil, err)
	assert(t, messageType == websocket.BinaryMessage)
	assert(t, string(bts) == message, "Expected:", message, "Received:", string(bts))
}
