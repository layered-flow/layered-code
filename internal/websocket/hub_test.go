package websocket

import (
	"testing"
	"time"
)

func TestHub(t *testing.T) {
	hub := NewHub()
	
	// Test hub creation
	if hub == nil {
		t.Fatal("Expected hub to be created")
	}
	
	// Start hub in background
	go hub.Run()
	
	// Create mock client
	mockClient := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	
	// Test registration
	hub.register <- mockClient
	time.Sleep(10 * time.Millisecond) // Allow goroutine to process
	
	// Test broadcast
	testMsg := []byte("test message")
	hub.broadcast <- testMsg
	
	// Check message received
	select {
	case msg := <-mockClient.send:
		if string(msg) != string(testMsg) {
			t.Errorf("Expected %s, got %s", testMsg, msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive broadcast message")
	}
	
	// Test unregistration
	hub.unregister <- mockClient
	time.Sleep(10 * time.Millisecond)
	
	// Verify client channel is closed
	select {
	case _, ok := <-mockClient.send:
		if ok {
			t.Error("Expected client channel to be closed")
		}
	default:
		// Channel not closed yet, wait a bit more
		time.Sleep(10 * time.Millisecond)
		select {
		case _, ok := <-mockClient.send:
			if ok {
				t.Error("Expected client channel to be closed")
			}
		default:
			t.Error("Client channel should be closed")
		}
	}
}