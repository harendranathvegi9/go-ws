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
	ws.UpgradeRequests("/ws/echo", func(event *ws.Event, conn *ws.Conn) {
		switch event.Type {
		case ws.Connected:
			log.Println("Client connected:", conn.RemoteAddr)

		case ws.TextMessage:
			text, err := event.Text()
			log.Println("Text message:", text, err)
			conn.SendText(text)

		case ws.BinaryMessage:
			data, err := event.Data() // Or use `event` as an `io.Reader`
			log.Println("Binary message size:", len(data), err)
			conn.SendBinary(data)

		case ws.Error:
			log.Println("Conn error:", event.Error)
			panic(event.Error)

		case ws.Disconnected:
			log.Println("Client disconnected:", conn.RemoteAddr)
		}
	})
	go http.ListenAndServe(addr, nil)
}

func TestServerAndClient(t *testing.T) {
	wsConn, _, err := dialer.Dial("ws://"+addr+"/ws/echo", nil)
	assert(t, err == nil, err)
	sendRecv(t, wsConn, websocket.BinaryMessage)
	sendRecv(t, wsConn, websocket.TextMessage)
	wsConn.Close()
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

func sendRecv(t *testing.T, ws *websocket.Conn, messageType int) {
	const message = "Hello World!"
	err := ws.SetWriteDeadline(time.Now().Add(time.Second))
	assert(t, err == nil, err)

	log.Println("Client: Send message", message)
	err = ws.WriteMessage(messageType, []byte(message))
	assert(t, err == nil, err)
	err = ws.SetReadDeadline(time.Now().Add(time.Second))
	assert(t, err == nil, err)

	log.Println("Client: Read next message")
	msgType, bts, err := ws.ReadMessage()
	log.Println("Client: Received message", message)
	assert(t, err == nil, err)
	assert(t, msgType == messageType)
	assert(t, string(bts) == message, "Expected:", message, "Received:", string(bts))
}
