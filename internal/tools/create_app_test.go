package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestCreateAppParams tests the CreateAppParams struct creation
func TestCreateAppParams(t *testing.T) {
	params := CreateAppParams{AppName: "testapp"}
	if params.AppName != "testapp" {
		t.Errorf("CreateAppParams not created correctly")
	}
}

// TestCreateAppResult tests the CreateAppResult struct creation
func TestCreateAppResult(t *testing.T) {
	result := CreateAppResult{Success: true, Message: "test message", Path: "/test/path"}
	if !result.Success || result.Message != "test message" || result.Path != "/test/path" {
		t.Errorf("CreateAppResult not created correctly")
	}
}

// TestValidateAppName tests the app name validation logic
func TestValidateAppName(t *testing.T) {
	tests := []struct {
		name    string
		appName string
		wantErr bool
	}{
		{"valid name", "myapp", false},
		{"valid name with numbers", "app123", false},
		{"valid name with dash", "my-app", false},
		{"valid name with underscore", "my_app", false},
		{"empty name", "", true},
		{"name with slash", "my/app", true},
		{"name with backslash", "my\\app", true},
		{"name with colon", "my:app", true},
		{"name with asterisk", "my*app", true},
		{"name with question mark", "my?app", true},
		{"name with quote", "my\"app", true},
		{"name with less than", "my<app", true},
		{"name with greater than", "my>app", true},
		{"name with pipe", "my|app", true},
		{"name with dot dot", "my..app", true},
		{"name with dot dot slash", "../app", true},
		{"hidden directory", ".myapp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAppName(tt.appName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAppName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCreateApp tests the core CreateApp functionality
func TestCreateApp(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	os.MkdirAll(appsDir, 0755)

	t.Run("create new app", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		params := CreateAppParams{AppName: "newapp"}
		result, err := CreateApp(params)
		if err != nil {
			t.Fatalf("CreateApp() failed: %v", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true, got false")
		}

		expectedPath := filepath.Join(appsDir, "newapp")
		if result.Path != expectedPath {
			t.Errorf("Expected path=%s, got %s", expectedPath, result.Path)
		}

		// Verify directory was created
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("App directory was not created")
		}

		// Verify subdirectories were created
		subdirs := []string{"src", "build", ".layered-code"}
		for _, subdir := range subdirs {
			subdirPath := filepath.Join(expectedPath, subdir)
			if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
				t.Errorf("Subdirectory %s was not created", subdir)
			}
		}

		// Verify files were created
		files := []string{
			".gitignore",
			".layered.json",
			"README.md",
			filepath.Join("src", "index.html"),
		}
		for _, file := range files {
			filePath := filepath.Join(expectedPath, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("File %s was not created", file)
			}
		}

		// Verify .gitignore content
		gitignoreContent, err := os.ReadFile(filepath.Join(expectedPath, ".gitignore"))
		if err != nil {
			t.Errorf("Failed to read .gitignore: %v", err)
		}
		expectedIgnores := []string{".layered-code/", "build/", "dist/", "node_modules/"}
		for _, ignore := range expectedIgnores {
			if !contains(string(gitignoreContent), ignore) {
				t.Errorf(".gitignore missing expected entry: %s", ignore)
			}
		}
	})

	t.Run("app already exists", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		// Create an app first
		existingApp := filepath.Join(appsDir, "existingapp")
		os.MkdirAll(existingApp, 0755)

		params := CreateAppParams{AppName: "existingapp"}
		result, err := CreateApp(params)
		if err == nil {
			t.Errorf("Expected error for existing app, got nil")
		}

		if result.Success {
			t.Errorf("Expected success=false for existing app")
		}
	})

	t.Run("invalid app name", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		params := CreateAppParams{AppName: "invalid/name"}
		result, err := CreateApp(params)
		if err == nil {
			t.Errorf("Expected error for invalid app name, got nil")
		}

		if result.Success {
			t.Errorf("Expected success=false for invalid app name")
		}
	})
}

// TestCreateAppMcp tests the MCP interface wrapper
func TestCreateAppMcp(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	os.MkdirAll(appsDir, 0755)

	t.Run("successful creation", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		ctx := context.Background()
		request := mcp.CallToolRequest{}
		request.Params.Name = "create_app"
		request.Params.Arguments = map[string]any{
			"app_name": "mcptestapp",
		}

		result, err := CreateAppMcp(ctx, request)
		if err != nil {
			t.Fatalf("CreateAppMcp() failed: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	t.Run("missing app_name", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		ctx := context.Background()
		request := mcp.CallToolRequest{}
		request.Params.Name = "create_app"
		request.Params.Arguments = map[string]any{}

		_, err := CreateAppMcp(ctx, request)
		if err == nil {
			t.Errorf("Expected error for missing app_name, got nil")
		}
	})
}

// TestCreateAppFunctionExecutions tests that all exported functions execute without panicking
func TestCreateAppFunctionExecutions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"CreateApp", func() error {
			_, err := CreateApp(CreateAppParams{AppName: "test"})
			return err
		}},
		{"CreateAppCli", func() error {
			// Save original args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"layered-code", "tool", "create_app", "testapp"}
			return CreateAppCli()
		}},
		{"CreateAppMcp", func() error {
			ctx := context.Background()
			request := mcp.CallToolRequest{}
			request.Params.Name = "create_app"
			request.Params.Arguments = map[string]any{
				"app_name": "test",
			}
			_, err := CreateAppMcp(ctx, request)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			_ = tt.fn() // Errors are expected due to filesystem/missing dirs
		})
	}
}