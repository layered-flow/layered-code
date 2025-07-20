package pnpm

import (
	"strings"
	"testing"
)

func TestPnpmViewValidation(t *testing.T) {
	tests := []struct {
		name         string
		packageName  string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "Empty package name",
			packageName:  "",
			expectError:  true,
			errorMessage: "package name is required",
		},
		{
			name:         "Whitespace only",
			packageName:  "   ",
			expectError:  true,
			errorMessage: "package name is required",
		},
		{
			name:         "Invalid characters - semicolon",
			packageName:  "package;name",
			expectError:  true,
			errorMessage: "invalid characters in package name",
		},
		{
			name:         "Invalid characters - pipe",
			packageName:  "package|name",
			expectError:  true,
			errorMessage: "invalid characters in package name",
		},
		{
			name:         "Invalid scoped package - no slash",
			packageName:  "@scope",
			expectError:  true,
			errorMessage: "invalid scoped package format",
		},
		{
			name:         "Invalid scoped package - empty after slash",
			packageName:  "@scope/",
			expectError:  true,
			errorMessage: "invalid scoped package format",
		},
		{
			name:         "Invalid scoped package - empty scope",
			packageName:  "@/package",
			expectError:  true,
			errorMessage: "invalid scoped package format",
		},
		{
			name:        "Valid simple package",
			packageName: "express",
			expectError: false,
		},
		{
			name:        "Valid package with version",
			packageName: "express@4.18.0",
			expectError: false,
		},
		{
			name:        "Valid scoped package",
			packageName: "@types/node",
			expectError: false,
		},
		{
			name:        "Valid scoped package with version",
			packageName: "@types/node@18.0.0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PnpmView(tt.packageName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorMessage, err)
				}
			} else {
				// Valid input but npm might not be installed or package might not exist
				// We only check that validation passes, not the full execution
				if err != nil && strings.Contains(err.Error(), "invalid") {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestPnpmViewResult(t *testing.T) {
	// Test that the struct can be properly populated
	result := PnpmViewResult{
		Name:        "test-package",
		Version:     "1.0.0",
		Description: "A test package",
		Homepage:    "https://example.com",
		License:     "MIT",
		Repository: map[string]interface{}{
			"type": "git",
			"url":  "https://github.com/user/repo.git",
		},
		Keywords: []string{"test", "package"},
		Dependencies: map[string]string{
			"express": "^4.18.0",
			"cors":    "^2.8.5",
		},
	}

	if result.Name != "test-package" {
		t.Errorf("Expected Name to be 'test-package', got '%s'", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got '%s'", result.Version)
	}
	if len(result.Keywords) != 2 {
		t.Errorf("Expected 2 keywords, got %d", len(result.Keywords))
	}
	if len(result.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(result.Dependencies))
	}
}