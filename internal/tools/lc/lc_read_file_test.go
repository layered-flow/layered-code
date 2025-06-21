package lc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestLcReadFile tests the core LcReadFile functionality including successful reads
// and various error conditions
func TestLcReadFile(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)

	// Create test files
	testContent := "package main"
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte(testContent), 0644)
	os.WriteFile(filepath.Join(appDir, "binary.bin"), []byte{0x00, 0xFF}, 0644)
	os.WriteFile(filepath.Join(appDir, "large.txt"), []byte(strings.Repeat("a", constants.MaxFileSize+1)), 0644)
	os.Symlink(filepath.Join(appDir, "main.go"), filepath.Join(appDir, "symlink.go"))

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	t.Run("successful read", func(t *testing.T) {
		result, err := LcReadFile("testapp", "main.go")
		if err != nil {
			t.Fatalf("ReadFile() failed: %v", err)
		}
		if result.AppName != "testapp" {
			t.Errorf("AppName = %s; want testapp", result.AppName)
		}
		if result.FilePath != "main.go" {
			t.Errorf("FilePath = %s; want main.go", result.FilePath)
		}
		if result.Content != testContent {
			t.Errorf("Content mismatch")
		}
	})

	t.Run("input validation errors", func(t *testing.T) {
		tests := []struct {
			appName, filePath, wantErr string
		}{
			{"", "main.go", "app_name is required"},
			{"testapp", "", "file_path is required"},
			{"testapp", "nonexistent.go", "no such file"},
		}
		for _, tt := range tests {
			_, err := LcReadFile(tt.appName, tt.filePath)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ReadFile(%q, %q) expected error containing %q, got: %v",
					tt.appName, tt.filePath, tt.wantErr, err)
			}
		}
	})

	t.Run("file type restrictions", func(t *testing.T) {
		tests := []struct {
			filePath string
			wantErr  error
		}{
			{"binary.bin", ErrBinaryFile},
			{"large.txt", ErrFileTooLarge},
			{"symlink.go", ErrSymlink},
		}
		for _, tt := range tests {
			_, err := LcReadFile("testapp", tt.filePath)
			if err != tt.wantErr {
				t.Errorf("ReadFile(testapp, %q) = %v; want %v", tt.filePath, err, tt.wantErr)
			}
		}
	})

	t.Run("path traversal attempt", func(t *testing.T) {
		_, err := LcReadFile("testapp", "../../../etc/passwd")
		if err == nil || !strings.Contains(err.Error(), "outside app directory") {
			t.Error("Expected error for path traversal attempt")
		}
	})
}

// TestLcReadFileCli tests the CLI interface
func TestLcReadFileCli(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "test.go"), []byte("package main"), 0644)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	// Save original os.Args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	t.Run("missing arguments", func(t *testing.T) {
		tests := []struct {
			args    []string
			wantErr string
		}{
			{[]string{"cmd", "tool", "read_file"}, "--app-name is required"},
			{[]string{"cmd", "tool", "read_file", "--app-name", "testapp"}, "--file-path is required"},
			{[]string{"cmd", "tool", "read_file", "--app-name"}, "--app-name requires a value"},
			{[]string{"cmd", "tool", "read_file", "--file-path"}, "--file-path requires a value"},
			{[]string{"cmd", "tool", "read_file", "--unknown"}, "unknown option: --unknown"},
		}
		for _, tt := range tests {
			os.Args = tt.args
			err := LcReadFileCli()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ReadFileCli() with args %v expected error containing %q, got: %v",
					tt.args[3:], tt.wantErr, err)
			}
		}
	})

	t.Run("help flag", func(t *testing.T) {
		for _, helpFlag := range []string{"--help", "-h"} {
			os.Args = []string{"cmd", "tool", "read_file", helpFlag}
			err := LcReadFileCli()
			if err != nil {
				t.Errorf("ReadFileCli() with %s should not error, got: %v", helpFlag, err)
			}
		}
	})

	t.Run("successful execution", func(t *testing.T) {
		os.Args = []string{"cmd", "tool", "read_file", "--app-name", "testapp", "--file-path", "test.go"}
		err := LcReadFileCli()
		if err != nil {
			t.Errorf("ReadFileCli() failed: %v", err)
		}
	})
}

// TestLcReadFileMcp tests the MCP interface wrapper
func TestLcReadFileMcp(t *testing.T) {
	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "read_file"
	request.Params.Arguments = map[string]any{
		"app_name":  "nonexistent",
		"file_path": "test.go",
	}

	_, err := LcReadFileMcp(ctx, request)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}
}
