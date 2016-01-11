// Package ws implements common server websocket requirements, such as ping/pong heartbeats and helpers for streaming messages.
//
// Example usage:
//
// 	// Implements an echo server with logging of connection events
// 	ws.UpgradeRequests("/ws/echo", func(event *ws.Event, connection *ws.Connection) {
// 		switch event.Type {
// 		case ws.Connected:
// 			log.Println("Client connected:", connection.RemoteAddr)
//
// 		case ws.TextMessage:
// 			text, err := event.Text()
// 			log.Println("Text message:", text, err)
// 			connection.SendText(text)
//
// 		case ws.BinaryMessage:
// 			data, err := event.Data() // or use `event` as an `io.Reader`
// 			log.Println("Binary message size:", len(data), err)
// 			connection.SendBinary(data)
//
// 		case ws.Error:
// 			log.Println("Connection error:", event.Error)
// 			panic(event.Error)
//
// 		case ws.Disconnected:
// 			log.Println("Client disconnected:", connection.RemoteAddr)
// 		}
// 	})
// 	go http.ListenAndServe(addr, nil) // ws.UpgradeRequests expects a listening http server

package ws
