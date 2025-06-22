package lc

import (
	"strings"
	"testing"
)

func TestPromptRead(t *testing.T) {
	tests := []struct {
		name         string
		promptName   string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Read by numeric ID",
			promptName:  "1",
			expectError: false,
		},
		{
			name:        "Read by name",
			promptName:  "General Principles",
			expectError: false,
		},
		{
			name:        "Read with lowercase",
			promptName:  "general principles",
			expectError: false,
		},
		{
			name:        "Read with different case",
			promptName:  "GENERAL PRINCIPLES",
			expectError: false,
		},
		{
			name:         "Read non-existent prompt",
			promptName:   "Non Existent Prompt",
			expectError:  true,
			errorMessage: "not found",
		},
		{
			name:         "Read empty prompt name",
			promptName:   "",
			expectError:  true,
			errorMessage: "not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PromptRead(tt.promptName)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorMessage, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				
				// Verify result has all required fields
				if result.Name != "General Principles" {
					t.Errorf("Expected name 'General Principles', got '%s'", result.Name)
				}
				
				if result.Description == "" {
					t.Error("Description should not be empty")
				}
				
				if result.Content == "" {
					t.Error("Content should not be empty")
				}
				
				// Verify content contains expected text
				if !strings.Contains(result.Content, "General Principles") {
					t.Error("Content does not contain expected text")
				}
			}
		})
	}
}

func TestPromptsMapIntegrity(t *testing.T) {
	// Verify all prompts in the map have valid data
	for key, prompt := range promptsMap {
		if prompt.ID == 0 {
			t.Errorf("Prompt with key %d has empty ID", key)
		}
		
		if prompt.Name == "" {
			t.Errorf("Prompt with key %d has empty name", key)
		}
		
		if prompt.Description == "" {
			t.Errorf("Prompt '%s' has empty description", prompt.Name)
		}
		
		if prompt.Content == "" {
			t.Errorf("Prompt '%s' has empty content", prompt.Name)
		}
		
		// Verify key matches ID
		if key != prompt.ID {
			t.Errorf("Key %d does not match prompt ID %d", key, prompt.ID)
		}
	}
}