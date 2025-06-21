package helpers

import (
	"testing"
)

func TestValidateAppName(t *testing.T) {
	tests := []struct {
		name        string
		appName     string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid app name",
			appName:     "my-app",
			shouldError: false,
		},
		{
			name:        "Empty app name",
			appName:     "",
			shouldError: true,
			errorMsg:    "app name cannot be empty",
		},
		{
			name:        "Path traversal attempt",
			appName:     "../evil",
			shouldError: true,
			errorMsg:    "app name cannot contain '..'",
		},
		{
			name:        "Just dots",
			appName:     "..",
			shouldError: true,
			errorMsg:    "app name cannot contain '..'",
		},
		{
			name:        "Contains forward slash",
			appName:     "my/app",
			shouldError: true,
			errorMsg:    "app name cannot contain '/'",
		},
		{
			name:        "Contains backslash",
			appName:     "my\\app",
			shouldError: true,
			errorMsg:    "app name cannot contain '\\'",
		},
		{
			name:        "Contains special characters",
			appName:     "my*app",
			shouldError: true,
			errorMsg:    "app name cannot contain '*'",
		},
		{
			name:        "Starts with period (hidden directory)",
			appName:     ".hidden-app",
			shouldError: true,
			errorMsg:    "app name cannot start with a period (hidden directories are not allowed)",
		},
		{
			name:        "Contains colon",
			appName:     "my:app",
			shouldError: true,
			errorMsg:    "app name cannot contain ':'",
		},
		{
			name:        "Contains question mark",
			appName:     "my?app",
			shouldError: true,
			errorMsg:    "app name cannot contain '?'",
		},
		{
			name:        "Contains quote",
			appName:     "my\"app",
			shouldError: true,
			errorMsg:    "app name cannot contain '\"'",
		},
		{
			name:        "Contains less than",
			appName:     "my<app",
			shouldError: true,
			errorMsg:    "app name cannot contain '<'",
		},
		{
			name:        "Contains greater than",
			appName:     "my>app",
			shouldError: true,
			errorMsg:    "app name cannot contain '>'",
		},
		{
			name:        "Contains pipe",
			appName:     "my|app",
			shouldError: true,
			errorMsg:    "app name cannot contain '|'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAppName(tt.appName)
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for app name '%s' but got none", tt.appName)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for app name '%s': %v", tt.appName, err)
			}
			if tt.shouldError && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}