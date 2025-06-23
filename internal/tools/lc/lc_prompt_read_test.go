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
			name:         "Read non-existent ID",
			promptName:   "999",
			expectError:  true,
			errorMessage: "not found",
		},
		{
			name:         "Read invalid ID",
			promptName:   "invalid",
			expectError:  true,
			errorMessage: "must be numeric",
		},
		{
			name:         "Read empty prompt name",
			promptName:   "",
			expectError:  true,
			errorMessage: "must be numeric",
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
			t.Errorf("Prompt with key %+v has empty ID", key)
		}
		
		if prompt.Version == 0 {
			t.Errorf("Prompt with key %+v has empty Version", key)
		}
		
		if prompt.Name == "" {
			t.Errorf("Prompt with key %+v has empty name", key)
		}
		
		if prompt.Description == "" {
			t.Errorf("Prompt '%s' has empty description", prompt.Name)
		}
		
		if prompt.Content == "" {
			t.Errorf("Prompt '%s' has empty content", prompt.Name)
		}
		
		// Verify key matches ID and Version
		if key.ID != prompt.ID || key.Version != prompt.Version {
			t.Errorf("Key %+v does not match prompt ID %d and Version %d", key, prompt.ID, prompt.Version)
		}
	}
}

func TestPromptReadWithVersion(t *testing.T) {
	tests := []struct {
		name         string
		promptName   string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Read by ID with version",
			promptName:  "1:1",
			expectError: false,
		},
		{
			name:         "Read non-existent version",
			promptName:   "1:99",
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
				
				// Verify result has expected fields
				if result.Name != "General Principles" {
					t.Errorf("Expected name 'General Principles', got '%s'", result.Name)
				}
			}
		})
	}
}

func TestPromptReadVersionEdgeCases(t *testing.T) {
	// Save original promptsMap
	originalMap := promptsMap
	defer func() { promptsMap = originalMap }()
	
	// Create test data with multiple versions
	promptsMap = map[PromptKey]Prompt{
		{ID: 1, Version: 1}: {
			ID:          1,
			Version:     1,
			Name:        "Test Prompt",
			Description: "Version 1",
			Content:     "Content v1",
		},
		{ID: 1, Version: 3}: {
			ID:          1,
			Version:     3,
			Name:        "Test Prompt",
			Description: "Version 3",
			Content:     "Content v3",
		},
		{ID: 1, Version: 5}: {
			ID:          1,
			Version:     5,
			Name:        "Test Prompt",
			Description: "Version 5",
			Content:     "Content v5",
		},
	}
	
	tests := []struct {
		name           string
		promptName     string
		expectError    bool
		expectedContent string
		errorMessage   string
	}{
		{
			name:           "Get latest version by ID",
			promptName:     "1",
			expectError:    false,
			expectedContent: "Content v5",
		},
		{
			name:           "Get specific existing version",
			promptName:     "1:3",
			expectError:    false,
			expectedContent: "Content v3",
		},
		{
			name:           "Get non-existent version",
			promptName:     "1:2",
			expectError:    true,
			errorMessage:   "not found",
		},
		{
			name:           "Get version 0",
			promptName:     "1:0",
			expectError:    true,
			errorMessage:   "not found",
		},
		{
			name:           "Invalid version format",
			promptName:     "1:abc",
			expectError:    false, // Will parse as ID 1, ignore invalid version and get latest
			expectedContent: "Content v5",
		},
		{
			name:           "Multiple colons",
			promptName:     "1:2:3",
			expectError:    false,  // Will parse as ID 1, version parsing fails so gets latest
			expectedContent: "Content v5",
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
				} else if tt.expectedContent != "" && result.Content != tt.expectedContent {
					t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, result.Content)
				}
			}
		})
	}
}