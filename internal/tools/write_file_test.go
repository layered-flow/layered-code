package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestWriteFile tests the core WriteFile functionality including successful writes
// and various error conditions
func TestWriteFile(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)
	
	// Create build directory
	outputDir := filepath.Join(appDir, constants.OutputDirectoryName)
	os.MkdirAll(outputDir, 0755)

	// Create an existing file for overwrite tests
	existingFile := filepath.Join(outputDir, "existing.txt")
	os.WriteFile(existingFile, []byte("old content"), 0644)

	// Create a directory for conflict tests
	os.MkdirAll(filepath.Join(outputDir, "testdir"), 0755)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	t.Run("successful create", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "new.txt",
			Content:  "new content",
			Mode:     "create",
		}
		result, err := WriteFile(params)
		if err != nil {
			t.Fatalf("WriteFile() failed: %v", err)
		}
		if result.AppName != "testapp" {
			t.Errorf("AppName = %s; want testapp", result.AppName)
		}
		if result.FilePath != "new.txt" {
			t.Errorf("FilePath = %s; want new.txt", result.FilePath)
		}
		if result.BytesWritten != len("new content") {
			t.Errorf("BytesWritten = %d; want %d", result.BytesWritten, len("new content"))
		}
		if !result.Created {
			t.Error("Created = false; want true")
		}

		// Verify file content
		content, _ := os.ReadFile(filepath.Join(outputDir, "new.txt"))
		if string(content) != "new content" {
			t.Errorf("File content = %q; want %q", content, "new content")
		}
	})

	t.Run("successful overwrite", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "existing.txt",
			Content:  "updated content",
			Mode:     "overwrite",
		}
		result, err := WriteFile(params)
		if err != nil {
			t.Fatalf("WriteFile() failed: %v", err)
		}
		if result.Created {
			t.Error("Created = true; want false")
		}

		// Verify file content
		content, _ := os.ReadFile(existingFile)
		if string(content) != "updated content" {
			t.Errorf("File content = %q; want %q", content, "updated content")
		}
	})

	t.Run("create with subdirectories", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "deep/nested/file.txt",
			Content:  "nested content",
			Mode:     "create",
		}
		result, err := WriteFile(params)
		if err != nil {
			t.Fatalf("WriteFile() failed: %v", err)
		}
		if !result.Created {
			t.Error("Created = false; want true")
		}

		// Verify file exists
		content, _ := os.ReadFile(filepath.Join(outputDir, "deep/nested/file.txt"))
		if string(content) != "nested content" {
			t.Errorf("File content = %q; want %q", content, "nested content")
		}
	})

	t.Run("input validation errors", func(t *testing.T) {
		tests := []struct {
			params  WriteFileParams
			wantErr string
		}{
			{WriteFileParams{FilePath: "test.txt", Content: "test"}, "app_name is required"},
			{WriteFileParams{AppName: "testapp", Content: "test"}, "file_path is required"},
			{WriteFileParams{AppName: "testapp", FilePath: "test.txt", Content: "test", Mode: "invalid"}, "invalid mode"},
			{WriteFileParams{AppName: "nonexistent", FilePath: "test.txt", Content: "test"}, "app directory does not exist"},
		}
		for _, tt := range tests {
			_, err := WriteFile(tt.params)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("WriteFile(%+v) expected error containing %q, got: %v",
					tt.params, tt.wantErr, err)
			}
		}
	})

	t.Run("create mode with existing file", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "existing.txt",
			Content:  "should fail",
			Mode:     "create",
		}
		_, err := WriteFile(params)
		if err == nil || !strings.Contains(err.Error(), "file already exists") {
			t.Errorf("Expected 'file already exists' error, got: %v", err)
		}
	})

	t.Run("write to directory", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "testdir",
			Content:  "should fail",
			Mode:     "overwrite",
		}
		_, err := WriteFile(params)
		if err == nil || !strings.Contains(err.Error(), "path is a directory") {
			t.Errorf("Expected 'path is a directory' error, got: %v", err)
		}
	})

	t.Run("file size limit", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "huge.txt",
			Content:  strings.Repeat("a", constants.MaxFileSize+1),
			Mode:     "create",
		}
		_, err := WriteFile(params)
		if err == nil || !strings.Contains(err.Error(), "exceeds maximum file size") {
			t.Errorf("Expected file size error, got: %v", err)
		}
	})

	t.Run("path traversal attempt", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "../../../etc/passwd",
			Content:  "malicious",
			Mode:     "overwrite",
		}
		_, err := WriteFile(params)
		if err == nil || !strings.Contains(err.Error(), "outside build directory") {
			t.Error("Expected error for path traversal attempt")
		}
	})

	t.Run("default mode", func(t *testing.T) {
		params := WriteFileParams{
			AppName:  "testapp",
			FilePath: "default-mode.txt",
			Content:  "test content",
			// Mode not specified, should default to "create"
		}
		result, err := WriteFile(params)
		if err != nil {
			t.Fatalf("WriteFile() failed: %v", err)
		}
		if !result.Created {
			t.Error("Created = false; want true")
		}
	})
}

// TestWriteFileCli tests the CLI interface
func TestWriteFileCli(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)
	
	// Create build directory
	outputDir := filepath.Join(appDir, constants.OutputDirectoryName)
	os.MkdirAll(outputDir, 0755)

	// Create a file with content for --content-file tests
	contentFile := filepath.Join(tempDir, "content.txt")
	os.WriteFile(contentFile, []byte("file content"), 0644)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	// Save original os.Args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	t.Run("missing arguments", func(t *testing.T) {
		tests := []struct {
			args    []string
			wantErr string
		}{
			{[]string{"cmd", "tool", "write_file"}, "--app-name is required"},
			{[]string{"cmd", "tool", "write_file", "--app-name", "testapp"}, "--file-path is required"},
			{[]string{"cmd", "tool", "write_file", "--app-name", "testapp", "--file-path", "test.txt"}, "either --content or --content-file is required"},
			{[]string{"cmd", "tool", "write_file", "--app-name"}, "--app-name requires a value"},
			{[]string{"cmd", "tool", "write_file", "--file-path"}, "--file-path requires a value"},
			{[]string{"cmd", "tool", "write_file", "--content"}, "--content requires a value"},
			{[]string{"cmd", "tool", "write_file", "--content-file"}, "--content-file requires a value"},
			{[]string{"cmd", "tool", "write_file", "--mode"}, "--mode requires a value"},
			{[]string{"cmd", "tool", "write_file", "--unknown"}, "unknown option: --unknown"},
		}
		for _, tt := range tests {
			os.Args = tt.args
			err := WriteFileCli()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("WriteFileCli() with args %v expected error containing %q, got: %v",
					tt.args[3:], tt.wantErr, err)
			}
		}
	})

	t.Run("both content options", func(t *testing.T) {
		os.Args = []string{"cmd", "tool", "write_file", "--app-name", "testapp", "--file-path", "test.txt",
			"--content", "inline", "--content-file", contentFile}
		err := WriteFileCli()
		if err == nil || !strings.Contains(err.Error(), "cannot use both --content and --content-file") {
			t.Errorf("Expected error for using both content options, got: %v", err)
		}
	})

	t.Run("help flag", func(t *testing.T) {
		for _, helpFlag := range []string{"--help", "-h"} {
			os.Args = []string{"cmd", "tool", "write_file", helpFlag}
			err := WriteFileCli()
			if err != nil {
				t.Errorf("WriteFileCli() with %s should not error, got: %v", helpFlag, err)
			}
		}
	})

	t.Run("successful with inline content", func(t *testing.T) {
		os.Args = []string{"cmd", "tool", "write_file", "--app-name", "testapp",
			"--file-path", "cli-test.txt", "--content", "test content"}
		err := WriteFileCli()
		if err != nil {
			t.Errorf("WriteFileCli() failed: %v", err)
		}

		// Verify file was created
		content, _ := os.ReadFile(filepath.Join(outputDir, "cli-test.txt"))
		if string(content) != "test content" {
			t.Errorf("File content = %q; want %q", content, "test content")
		}
	})

	t.Run("successful with content file", func(t *testing.T) {
		os.Args = []string{"cmd", "tool", "write_file", "--app-name", "testapp",
			"--file-path", "cli-file-test.txt", "--content-file", contentFile}
		err := WriteFileCli()
		if err != nil {
			t.Errorf("WriteFileCli() failed: %v", err)
		}

		// Verify file was created
		content, _ := os.ReadFile(filepath.Join(outputDir, "cli-file-test.txt"))
		if string(content) != "file content" {
			t.Errorf("File content = %q; want %q", content, "file content")
		}
	})

	t.Run("with mode option", func(t *testing.T) {
		// Create file first
		os.Args = []string{"cmd", "tool", "write_file", "--app-name", "testapp",
			"--file-path", "mode-test.txt", "--content", "initial"}
		WriteFileCli()

		// Try to overwrite
		os.Args = []string{"cmd", "tool", "write_file", "--app-name", "testapp",
			"--file-path", "mode-test.txt", "--content", "updated", "--mode", "overwrite"}
		err := WriteFileCli()
		if err != nil {
			t.Errorf("WriteFileCli() with overwrite failed: %v", err)
		}

		// Verify content was updated
		content, _ := os.ReadFile(filepath.Join(outputDir, "mode-test.txt"))
		if string(content) != "updated" {
			t.Errorf("File content = %q; want %q", content, "updated")
		}
	})
}

// TestWriteFileMcp tests the MCP interface wrapper
func TestWriteFileMcp(t *testing.T) {
	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "write_file"
	request.Params.Arguments = map[string]any{
		"app_name":  "nonexistent",
		"file_path": "test.txt",
		"content":   "test content",
		"mode":      "create",
	}

	_, err := WriteFileMcp(ctx, request)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}
}