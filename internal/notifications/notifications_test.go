package notifications

import (
	"testing"

	"github.com/layered-flow/layered-code/internal/websocket"
)

func TestSetHub(t *testing.T) {
	// Test nil hub
	SetHub(nil)
	if hub != nil {
		t.Error("Expected hub to be nil")
	}
	
	// Test valid hub
	testHub := &websocket.Hub{}
	SetHub(testHub)
	if hub != testHub {
		t.Error("Expected hub to be set")
	}
}

func TestNotifyFileChange(t *testing.T) {
	// Test with nil hub (should not panic)
	SetHub(nil)
	NotifyFileChange("test.txt", "edit")
	
	// Test with mock hub would require mocking Hub.NotifyFileChange
	// which is already tested in websocket package
}