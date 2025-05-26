package notifications

import "github.com/layered-flow/layered-code/internal/websocket"

var hub *websocket.Hub

// SetHub sets the WebSocket hub for notifications
func SetHub(h *websocket.Hub) {
	hub = h
}

// NotifyFileChange sends a file change notification if hub is available
func NotifyFileChange(filename string, action string) {
	if hub != nil {
		hub.NotifyFileChange(filename, action)
	}
}