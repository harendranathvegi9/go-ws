package ws_test

import (
	"fmt"
	"net/http"

	"github.com/marcuswestin/go-ws"
)

func ExampleUpgradeRequests_server() {
	// Implement an echo server with logging of connection events
	ws.UpgradeRequests("/ws/echo", func(event *ws.Event, conn ws.Conn) {
		switch event.Type {
		case ws.Connected:
			fmt.Println("Client connected:", conn.HTTPRequest().RemoteAddr)

		case ws.TextMessage:
			text, err := event.Text()
			fmt.Println("Echo message:", text, err)
			conn.SendText(text)

		case ws.BinaryMessage:
			data, err := event.Data() // or use `event` as an `io.Reader`
			fmt.Println("Echo binary message of size:", len(data), err)
			conn.SendBinary(data)

		case ws.Error:
			panic("Error! " + event.Error.Error())

		case ws.NetError:
			fmt.Println("Network error:", event.Error)

		case ws.Disconnected:
			fmt.Println("Client disconnected:", conn.HTTPRequest().RemoteAddr)
		}
	})
	// ws.UpgradeRequests requires a listening http server
	go http.ListenAndServe("localhost:8080", nil)
}

func ExampleConnect_client() {
	doneChan := make(chan bool)
	defer func() { <-doneChan }()

	ws.Connect("http://localhost:8087/ws/echo", func(event *ws.Event, conn ws.Conn) {
		switch event.Type {
		case ws.Connected:
			fmt.Println("- Connected")
			conn.SendText("Hello!")

		case ws.TextMessage:
			text, _ := event.Text()
			fmt.Println("- Received", text)
			conn.SendText("Disconnect me")

		case ws.Disconnected:
			fmt.Println("- Disconnected")
			doneChan <- true
		}
	})

	// Output:
	// - Connected
	// - Received Hello!
	// - Disconnected
}
