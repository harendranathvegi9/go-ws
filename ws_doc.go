// Package ws implements common websocket functionality, such as heartbeats and message data streaming.
//
// Event Handlers
//
// You use this packet primarily by specifying an event handler function, which then gets passed
// a series of events, such as Connected, TextMessage, and Disconnected.
//
// You can access the event's data either by reading it all right into memory using Event.Text() and Event.Data(),
// or you can stream the event's data by using it as an io.Reader. Note that all the underlying event data must be
// read during the duration of its EventHandler function invocation.
//
// Servers
//
// See UpgradeRequests() and its example for server usage.
//
// Clients
//
// See Connect() and its example for client usage.
package ws
