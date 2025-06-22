package lc

import (
	"testing"
)

func TestPromptList(t *testing.T) {
	result, err := PromptList()
	if err != nil {
		t.Fatalf("PromptList() returned error: %v", err)
	}
	
	// Check that we have at least one prompt
	if result.Count == 0 {
		t.Error("Expected at least one prompt, got 0")
	}
	
	// Check that General Principles is in the list
	found := false
	for _, prompt := range result.Prompts {
		if prompt.Name == "General Principles" {
			found = true
			// Verify the prompt has expected fields
			if prompt.ID != "general_principles" {
				t.Errorf("Expected ID 'general_principles', got '%s'", prompt.ID)
			}
			if prompt.Description == "" {
				t.Error("General Principles prompt has empty description")
			}
			// List result should not include content
			if prompt.Content != "" {
				t.Error("List result should not include content")
			}
			break
		}
	}
	
	if !found {
		t.Error("General Principles prompt not found in list")
	}
	
	// Verify count matches number of prompts
	if result.Count != len(result.Prompts) {
		t.Errorf("Count (%d) does not match number of prompts (%d)", result.Count, len(result.Prompts))
	}
}