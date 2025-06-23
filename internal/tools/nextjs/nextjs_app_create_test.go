package nextjs

import (
	"strings"
	"testing"
)

func TestNextjsAppCreateValidation(t *testing.T) {
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
			errMsg:    "app name cannot be empty",
		},
		{
			name:      "app name with directory traversal",
			appName:   "../evil",
			wantErr:   true,
			errMsg:    "app name cannot contain '..'",
		},
		{
			name:      "app name with forward slash",
			appName:   "path/to/app",
			wantErr:   true,
			errMsg:    "app name cannot contain '/'",
		},
		{
			name:      "app name with backslash",
			appName:   "path\\to\\app",
			wantErr:   true,
			errMsg:    "app name cannot contain '\\'",
		},
		{
			name:      "app name with dots only",
			appName:   "..",
			wantErr:   true,
			errMsg:    "app name cannot contain '..'",
		},
		{
			name:      "app name starting with period",
			appName:   ".hidden",
			wantErr:   true,
			errMsg:    "app name cannot start with a period (hidden directories are not allowed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the function - it will fail early on validation
			_, err := NextjsAppCreate(tt.appName, "", false)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("NextjsAppCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NextjsAppCreate() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestNextjsAppCreateTemplateValidation(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		template string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid template - typescript",
			appName:  "test-app",
			template: "typescript",
			wantErr:  false,
		},
		{
			name:     "valid template - javascript",
			appName:  "test-app",
			template: "javascript",
			wantErr:  false,
		},
		{
			name:     "valid template - tailwind",
			appName:  "test-app",
			template: "tailwind",
			wantErr:  false,
		},
		{
			name:     "valid template - app",
			appName:  "test-app",
			template: "app",
			wantErr:  false,
		},
		{
			name:     "valid template - app-tw",
			appName:  "test-app",
			template: "app-tw",
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
			_, err := NextjsAppCreate(tt.appName, tt.template, false)

			// For valid cases, we expect it to fail later (no package manager available in test)
			if !tt.wantErr && err != nil {
				// Check if it's failing for the right reason (package manager not found)
				if !strings.Contains(err.Error(), "neither pnpm nor npm is available") &&
				   !strings.Contains(err.Error(), "app") {
					t.Errorf("NextjsAppCreate() unexpected error = %v", err)
				}
				return
			}

			// Check error expectations for invalid cases
			if tt.wantErr && err == nil {
				t.Errorf("NextjsAppCreate() expected error but got none")
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NextjsAppCreate() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}