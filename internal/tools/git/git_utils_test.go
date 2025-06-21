package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/constants"
)

func TestValidateAppPath(t *testing.T) {
	// Create a temporary home directory for testing
	tempHome, err := os.MkdirTemp("", "test-home")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	// Save current env and restore after test
	oldHome := os.Getenv("HOME")
	oldAppsDir := os.Getenv(constants.AppsDirectoryEnvVar)
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv(constants.AppsDirectoryEnvVar, oldAppsDir)
	}()

	// Set test environment
	os.Setenv("HOME", tempHome)
	os.Unsetenv(constants.AppsDirectoryEnvVar) // Use default

	// Create the default apps directory
	appsDir := filepath.Join(tempHome, constants.DefaultAppsDirectory)
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		appPath     string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "valid app path",
			appPath:     filepath.Join(appsDir, "myapp"),
			shouldError: false,
		},
		{
			name:        "path with traversal attempt",
			appPath:     filepath.Join(appsDir, "../outside"),
			shouldError: true,
			errorMsg:    "potential directory traversal detected",
		},
		{
			name:        "path outside apps directory",
			appPath:     "/tmp/malicious",
			shouldError: true,
			errorMsg:    "potential directory traversal detected",
		},
		{
			name:        "path with multiple traversals",
			appPath:     filepath.Join(appsDir, "app1", "..", "..", ".."),
			shouldError: true,
			errorMsg:    "potential directory traversal detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAppPath(tt.appPath)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none for path: %s", tt.appPath)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v for path: %s", err, tt.appPath)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}