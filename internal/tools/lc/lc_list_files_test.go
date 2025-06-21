package lc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestLcListFiles tests the core LcListFiles functionality including basic listing,
// pattern matching, and error handling
func TestLcListFiles(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)

	// Create test structure
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(appDir, "README.md"), []byte("# Test"), 0644)
	os.WriteFile(filepath.Join(appDir, ".hidden"), []byte("hidden"), 0644)
	os.Mkdir(filepath.Join(appDir, "src"), 0755)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	t.Run("basic listing", func(t *testing.T) {
		result, err := LcListFiles("testapp", nil, false, false, false)
		if err != nil {
			t.Fatalf("LcListFiles() failed: %v", err)
		}
		if result.AppName != "testapp" {
			t.Errorf("AppName = %s; want testapp", result.AppName)
		}
		// Should find root, main.go, README.md, src (but not .hidden)
		if len(result.Files) != 4 {
			t.Errorf("Found %d files; want 4", len(result.Files))
		}
	})

	t.Run("pattern matching", func(t *testing.T) {
		pattern := "*.go"
		result, err := LcListFiles("testapp", &pattern, false, false, false)
		if err != nil {
			t.Fatalf("LcListFiles() failed: %v", err)
		}
		// Should find only main.go
		fileCount := 0
		for _, file := range result.Files {
			if !file.IsDirectory {
				fileCount++
			}
		}
		if fileCount != 1 {
			t.Errorf("Pattern *.go found %d files; want 1", fileCount)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		if _, err := LcListFiles("", nil, false, false, false); err == nil {
			t.Error("Expected error for empty app name")
		}
		if _, err := LcListFiles("nonexistent", nil, false, false, false); err == nil {
			t.Error("Expected error for non-existent app")
		}
	})
}

// TestLcListFilesMcp tests the MCP interface wrapper to ensure it properly
// handles requests and returns appropriate errors
func TestLcListFilesMcp(t *testing.T) {
	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "lc_list_files"
	request.Params.Arguments = map[string]any{"app_name": "nonexistent"}

	_, err := LcListFilesMcp(ctx, request)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}
}

// TestFormatSize tests the utility function that converts byte counts
// into human-readable size strings
func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tt := range tests {
		if got := formatSize(tt.bytes); got != tt.expected {
			t.Errorf("formatSize(%d) = %s; want %s", tt.bytes, got, tt.expected)
		}
	}
}

// TestGetChildCount tests the utility function that counts non-hidden
// files and directories within a given directory
func TestGetChildCount(t *testing.T) {
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".hidden"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tempDir, "dir"), 0755)

	if got := getChildCount(tempDir); got != 2 {
		t.Errorf("getChildCount() = %d; want 2", got)
	}
}
