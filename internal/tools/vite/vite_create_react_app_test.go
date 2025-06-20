package vite

import (
	"strings"
	"testing"
)

func TestViteCreateReactAppValidation(t *testing.T) {
	// Test cases for validation only
	tests := []struct {
		name      string
		appName   string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty app name",
			appName:   "",
			wantErr:   true,
			errMsg:    "app name is required",
		},
		{
			name:      "app name with directory traversal",
			appName:   "../evil",
			wantErr:   true,
			errMsg:    "app name cannot contain path separators or '..'",
		},
		{
			name:      "app name with forward slash",
			appName:   "path/to/app",
			wantErr:   true,
			errMsg:    "app name cannot contain path separators or '..'",
		},
		{
			name:      "app name with backslash",
			appName:   "path\\to\\app",
			wantErr:   true,
			errMsg:    "app name cannot contain path separators or '..'",
		},
		{
			name:      "app name with dots only",
			appName:   "..",
			wantErr:   true,
			errMsg:    "app name cannot contain path separators or '..'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the function - it will fail early on validation
			_, err := ViteCreateReactApp(tt.appName, false)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("ViteCreateReactApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ViteCreateReactApp() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}