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

func TestPnpmPm2WithFlags(t *testing.T) {
	// Since we can't easily test the full PM2 execution without PM2 installed,
	// we'll test the command building logic by checking error messages
	tests := []struct {
		name          string
		command       string
		target        string
		flags         []string
		expectedError string
	}{
		{
			name:          "logs command without flags",
			command:       "logs",
			target:        "",
			flags:         []string{},
			expectedError: "pm2 logs --nostream",
		},
		{
			name:          "logs command with target",
			command:       "logs", 
			target:        "myapp",
			flags:         []string{},
			expectedError: "pm2 logs myapp --nostream",
		},
		{
			name:          "logs command with additional flags",
			command:       "logs",
			target:        "myapp",
			flags:         []string{"--lines", "100"},
			expectedError: "pm2 logs myapp --nostream --lines 100",
		},
		{
			name:          "logs command with multiple flags",
			command:       "logs",
			target:        "",
			flags:         []string{"--lines", "50", "--err"},
			expectedError: "pm2 logs --nostream --lines 50 --err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call PnpmPm2 which may succeed or fail depending on PM2 availability
			result, err := PnpmPm2(tt.command, tt.target, tt.flags, false)
			
			if err != nil {
				// If there's an error, it should contain the command that was attempted
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.expectedError, err)
				}
			} else {
				// If successful, verify the command was built correctly
				if !strings.Contains(result.Command, tt.expectedError) {
					t.Errorf("Expected command to contain '%s', but got: %s", tt.expectedError, result.Command)
				}
			}
		})
	}
}

