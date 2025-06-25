package lc

import (
	"testing"

	"github.com/layered-flow/layered-code/internal/tools/lc/shared"
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
			if prompt.ID != 1 {
				t.Errorf("Expected ID 1, got %d", prompt.ID)
			}
			if prompt.Version != 2 {
				t.Errorf("Expected Version 2 (highest version), got %d", prompt.Version)
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

func TestPromptListWithMultipleVersions(t *testing.T) {
	// Save original promptsMap
	originalMap := shared.PromptsMap
	defer func() { shared.PromptsMap = originalMap }()
	
	// Create test data with multiple versions
	shared.PromptsMap = map[shared.PromptKey]shared.Prompt{
		{ID: 1, Version: 1}: {
			ID:          1,
			Version:     1,
			Name:        "Test Prompt",
			Description: "Version 1",
			Content:     "Content v1",
		},
		{ID: 1, Version: 2}: {
			ID:          1,
			Version:     2,
			Name:        "Test Prompt",
			Description: "Version 2",
			Content:     "Content v2",
		},
		{ID: 1, Version: 3}: {
			ID:          1,
			Version:     3,
			Name:        "Test Prompt",
			Description: "Version 3",
			Content:     "Content v3",
		},
		{ID: 2, Version: 1}: {
			ID:          2,
			Version:     1,
			Name:        "Another Prompt",
			Description: "Only version",
			Content:     "Content",
		},
	}
	
	result, err := PromptList()
	if err != nil {
		t.Fatalf("PromptList() returned error: %v", err)
	}
	
	// Should only return 2 prompts (highest version of each)
	if result.Count != 2 {
		t.Errorf("Expected 2 prompts, got %d", result.Count)
	}
	
	// Check that we got the highest versions
	for _, prompt := range result.Prompts {
		if prompt.ID == 1 && prompt.Version != 3 {
			t.Errorf("Expected version 3 for prompt ID 1, got %d", prompt.Version)
		}
		if prompt.ID == 2 && prompt.Version != 1 {
			t.Errorf("Expected version 1 for prompt ID 2, got %d", prompt.Version)
		}
	}
}