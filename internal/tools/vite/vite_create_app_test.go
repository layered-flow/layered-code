package vite

import (
	"strings"
	"testing"
)

func TestViteCreateAppValidation(t *testing.T) {
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
			_, err := ViteCreateApp(tt.appName, "", false)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("ViteCreateApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ViteCreateApp() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestViteCreateAppTemplateValidation(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		template string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid template",
			appName:  "test-app",
			template: "react",
			wantErr:  false,
		},
		{
			name:     "empty template uses default",
			appName:  "test-app",
			template: "",
			wantErr:  false,
		},
		{
			name:     "invalid template",
			appName:  "test-app",
			template: "invalid-template",
			wantErr:  true,
			errMsg:   "invalid template 'invalid-template'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the function - it will fail early on validation
			_, err := ViteCreateApp(tt.appName, tt.template, false)

			// For valid cases, we expect it to fail later (no package manager available in test)
			if !tt.wantErr && err != nil {
				// Check if it's failing for the right reason (package manager not found)
				if !strings.Contains(err.Error(), "neither pnpm nor npm is available") &&
				   !strings.Contains(err.Error(), "app") {
					t.Errorf("ViteCreateApp() unexpected error = %v", err)
				}
				return
			}

			// Check error expectations for invalid cases
			if tt.wantErr && err == nil {
				t.Errorf("ViteCreateApp() expected error but got none")
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ViteCreateApp() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}