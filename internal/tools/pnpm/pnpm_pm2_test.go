package pnpm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetScriptToRun(t *testing.T) {
	tests := []struct {
		name         string
		packageJSON  string
		expectedScript string
		shouldError  bool
		createIndexJS bool
	}{
		{
			name: "Dev script exists",
			packageJSON: `{
				"scripts": {
					"dev": "vite",
					"start": "node server.js"
				}
			}`,
			expectedScript: "dev",
			shouldError: false,
			createIndexJS: false,
		},
		{
			name: "Only start script exists",
			packageJSON: `{
				"scripts": {
					"start": "node server.js"
				}
			}`,
			expectedScript: "start",
			shouldError: false,
			createIndexJS: false,
		},
		{
			name: "No scripts but index.js exists",
			packageJSON: `{}`,
			expectedScript: "index.js",
			shouldError: false,
			createIndexJS: true,
		},
		{
			name: "No suitable script",
			packageJSON: `{
				"scripts": {
					"test": "jest"
				}
			}`,
			expectedScript: "",
			shouldError: true,
			createIndexJS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for each test
			tmpDir, err := os.MkdirTemp("", "pm2-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Write package.json
			packagePath := filepath.Join(tmpDir, "package.json")
			if err := os.WriteFile(packagePath, []byte(tt.packageJSON), 0644); err != nil {
				t.Fatal(err)
			}

			// Create index.js if requested
			if tt.createIndexJS {
				indexPath := filepath.Join(tmpDir, "index.js")
				if err := os.WriteFile(indexPath, []byte("// test"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			script, err := getScriptToRun(packagePath)
			
			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if script != tt.expectedScript {
				t.Errorf("Expected script '%s', got '%s'", tt.expectedScript, script)
			}
		})
	}
}

func TestPnpmPm2Result(t *testing.T) {
	// Test result formatting
	result := PnpmPm2Result{
		AppName:        "test-app",
		AppPath:        "/path/to/test-app",
		PackageManager: "pnpm",
		Command:        "pnpm dlx pm2 list",
		Output:         "process list output",
		ErrorOutput:    "some warnings",
		Message:        "Successfully executed: pnpm dlx pm2 list",
	}

	// Check that all fields are properly set
	if result.AppName != "test-app" {
		t.Errorf("Expected AppName to be 'test-app', got '%s'", result.AppName)
	}
	if result.AppPath != "/path/to/test-app" {
		t.Errorf("Expected AppPath to be '/path/to/test-app', got '%s'", result.AppPath)
	}
	if result.PackageManager != "pnpm" {
		t.Errorf("Expected PackageManager to be 'pnpm', got '%s'", result.PackageManager)
	}
	if !strings.Contains(result.Command, "pm2 list") {
		t.Errorf("Expected Command to contain 'pm2 list', got '%s'", result.Command)
	}
	if result.Output != "process list output" {
		t.Errorf("Expected Output to be 'process list output', got '%s'", result.Output)
	}
	if result.ErrorOutput != "some warnings" {
		t.Errorf("Expected ErrorOutput to be 'some warnings', got '%s'", result.ErrorOutput)
	}
	if !strings.Contains(result.Message, "Successfully executed") {
		t.Errorf("Expected Message to contain 'Successfully executed', got '%s'", result.Message)
	}
}

func TestValidateAppName(t *testing.T) {
	tests := []struct {
		name        string
		appName     string
		shouldError bool
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
		},
		{
			name:        "Path traversal attempt",
			appName:     "../evil",
			shouldError: true,
		},
		{
			name:        "Contains forward slash",
			appName:     "my/app",
			shouldError: true,
		},
		{
			name:        "Contains backslash",
			appName:     "my\\app",
			shouldError: true,
		},
		{
			name:        "Contains special characters",
			appName:     "my*app",
			shouldError: true,
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
		})
	}
}