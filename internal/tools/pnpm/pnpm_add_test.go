package pnpm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPnpmAddValidation(t *testing.T) {
	tests := []struct {
		name         string
		appName      string
		packages     []string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "Empty app name",
			appName:      "",
			packages:     []string{"express"},
			expectError:  true,
			errorMessage: "app name is required",
		},
		{
			name:         "Invalid app name with slash",
			appName:      "my/app",
			packages:     []string{"express"},
			expectError:  true,
			errorMessage: "invalid app name",
		},
		{
			name:         "No packages provided",
			appName:      "myapp",
			packages:     []string{},
			expectError:  true,
			errorMessage: "at least one package name is required",
		},
		{
			name:         "Empty package in list",
			appName:      "myapp",
			packages:     []string{"express", "", "cors"},
			expectError:  true,
			errorMessage: "empty package name provided",
		},
		{
			name:         "Non-existent app",
			appName:      "nonexistentapp",
			packages:     []string{"express"},
			expectError:  true,
			errorMessage: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary apps directory within home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				t.Fatal(err)
			}
			tmpDir, err := os.MkdirTemp(homeDir, "pnpm-add-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Set LAYERED_APPS_DIRECTORY to our temp dir
			os.Setenv("LAYERED_APPS_DIRECTORY", tmpDir)
			defer os.Unsetenv("LAYERED_APPS_DIRECTORY")

			_, err = PnpmAdd(tt.appName, tt.packages, false)

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
			}
		})
	}
}

func TestPnpmAddCommandBuilding(t *testing.T) {
	// Create a temporary directory within home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir, err := os.MkdirTemp(homeDir, "pnpm-add-cmd-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an app directory with package.json
	appName := "testapp"
	appPath := filepath.Join(tmpDir, appName)
	if err := os.MkdirAll(appPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal package.json
	packageJSON := `{"name": "testapp", "version": "1.0.0"}`
	if err := os.WriteFile(filepath.Join(appPath, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Set LAYERED_APPS_DIRECTORY to our temp dir
	os.Setenv("LAYERED_APPS_DIRECTORY", tmpDir)
	defer os.Unsetenv("LAYERED_APPS_DIRECTORY")

	tests := []struct {
		name             string
		packages         []string
		expectedInError  string
		expectedInResult string
	}{
		{
			name:             "Single package",
			packages:         []string{"express"},
			expectedInError:  "add express",
			expectedInResult: "express",
		},
		{
			name:             "Multiple packages",
			packages:         []string{"express", "cors", "@types/node"},
			expectedInError:  "add express cors @types/node",
			expectedInResult: "express, cors, @types/node",
		},
		{
			name:             "Scoped package",
			packages:         []string{"@angular/core"},
			expectedInError:  "add @angular/core",
			expectedInResult: "@angular/core",
		},
		{
			name:             "Package with version",
			packages:         []string{"react@18.2.0"},
			expectedInError:  "add react@18.2.0",
			expectedInResult: "react@18.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will either succeed or fail depending on whether pnpm/npm is installed
			result, err := PnpmAdd(appName, tt.packages, false)
			
			if err != nil {
				// If there's an error, it should contain the command that was attempted
				if !strings.Contains(err.Error(), "add") || !strings.Contains(err.Error(), tt.packages[0]) {
					t.Errorf("Expected error to contain 'add' and package name, but got: %v", err)
				}
			} else {
				// If successful, check the result
				if !strings.Contains(result.Message, tt.expectedInResult) {
					t.Errorf("Expected result message to contain '%s', but got: %s", tt.expectedInResult, result.Message)
				}
				if len(result.Packages) != len(tt.packages) {
					t.Errorf("Expected %d packages in result, got %d", len(tt.packages), len(result.Packages))
				}
			}
		})
	}
}

func TestPnpmAddResult(t *testing.T) {
	// Test result formatting
	packages := []string{"express", "cors", "@types/node"}
	result := PnpmAddResult{
		AppName:        "myapp",
		AppPath:        "/path/to/myapp",
		PackageManager: "pnpm",
		Packages:       packages,
		Message:        "Successfully added express, cors, @types/node to 'myapp' using pnpm",
		Output:         "installation output",
		ErrorOutput:    "some warnings",
	}

	// Check that all fields are properly set
	if result.AppName != "myapp" {
		t.Errorf("Expected AppName to be 'myapp', got '%s'", result.AppName)
	}
	if result.AppPath != "/path/to/myapp" {
		t.Errorf("Expected AppPath to be '/path/to/myapp', got '%s'", result.AppPath)
	}
	if result.PackageManager != "pnpm" {
		t.Errorf("Expected PackageManager to be 'pnpm', got '%s'", result.PackageManager)
	}
	if len(result.Packages) != 3 {
		t.Errorf("Expected 3 packages, got %d", len(result.Packages))
	}
	if !strings.Contains(result.Message, "Successfully added") {
		t.Errorf("Expected Message to contain 'Successfully added', got '%s'", result.Message)
	}
	if !strings.Contains(result.Message, "express, cors, @types/node") {
		t.Errorf("Expected Message to contain all packages, got '%s'", result.Message)
	}
	if result.Output != "installation output" {
		t.Errorf("Expected Output to be 'installation output', got '%s'", result.Output)
	}
	if result.ErrorOutput != "some warnings" {
		t.Errorf("Expected ErrorOutput to be 'some warnings', got '%s'", result.ErrorOutput)
	}
}