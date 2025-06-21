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

// TestLcEditFile tests the core LcEditFile functionality
func TestLcEditFile(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	t.Run("successful replace all", func(t *testing.T) {
		testFile := filepath.Join(appDir, "test1.txt")
		os.WriteFile(testFile, []byte("hello world\nhello universe\nhello cosmos"), 0644)

		params := LcEditFileParams{
			AppName:     "testapp",
			FilePath:    "test1.txt",
			OldString:   "hello",
			NewString:   "goodbye",
			Occurrences: 0, // Replace all
		}
		result, err := LcEditFile(params)
		if err != nil {
			t.Fatalf("EditFile() failed: %v", err)
		}
		if result.Replacements != 3 {
			t.Errorf("Replacements = %d; want 3", result.Replacements)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		expected := "goodbye world\ngoodbye universe\ngoodbye cosmos"
		if string(content) != expected {
			t.Errorf("File content = %q; want %q", content, expected)
		}
	})

	t.Run("successful replace limited occurrences", func(t *testing.T) {
		testFile := filepath.Join(appDir, "test2.txt")
		os.WriteFile(testFile, []byte("foo bar foo baz foo"), 0644)

		params := LcEditFileParams{
			AppName:     "testapp",
			FilePath:    "test2.txt",
			OldString:   "foo",
			NewString:   "qux",
			Occurrences: 2,
		}
		result, err := LcEditFile(params)
		if err != nil {
			t.Fatalf("EditFile() failed: %v", err)
		}
		if result.Replacements != 2 {
			t.Errorf("Replacements = %d; want 2", result.Replacements)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		expected := "qux bar qux baz foo"
		if string(content) != expected {
			t.Errorf("File content = %q; want %q", content, expected)
		}
	})

	t.Run("delete text (empty new_string)", func(t *testing.T) {
		testFile := filepath.Join(appDir, "test3.txt")
		os.WriteFile(testFile, []byte("TODO: implement this\nTODO: fix that"), 0644)

		params := LcEditFileParams{
			AppName:     "testapp",
			FilePath:    "test3.txt",
			OldString:   "TODO: ",
			NewString:   "",
			Occurrences: 0,
		}
		result, err := LcEditFile(params)
		if err != nil {
			t.Fatalf("EditFile() failed: %v", err)
		}
		if result.Replacements != 2 {
			t.Errorf("Replacements = %d; want 2", result.Replacements)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		expected := "implement this\nfix that"
		if string(content) != expected {
			t.Errorf("File content = %q; want %q", content, expected)
		}
	})

	t.Run("preserve file content with special characters", func(t *testing.T) {
		testFile := filepath.Join(appDir, "test4.txt")
		originalContent := "line1\nline2\ttab\nline3\r\nline4"
		os.WriteFile(testFile, []byte(originalContent), 0644)

		params := LcEditFileParams{
			AppName:     "testapp",
			FilePath:    "test4.txt",
			OldString:   "line2",
			NewString:   "LINE2",
			Occurrences: 0,
		}
		_, err := LcEditFile(params)
		if err != nil {
			t.Fatalf("EditFile() failed: %v", err)
		}

		// Verify content preserves special characters
		content, _ := os.ReadFile(testFile)
		expected := "line1\nLINE2\ttab\nline3\r\nline4"
		if string(content) != expected {
			t.Errorf("File content = %q; want %q", content, expected)
		}
	})

	t.Run("input validation errors", func(t *testing.T) {
		tests := []struct {
			params  LcEditFileParams
			wantErr string
		}{
			{LcEditFileParams{FilePath: "test.txt", OldString: "old", NewString: "new"}, "app_name is required"},
			{LcEditFileParams{AppName: "testapp", OldString: "old", NewString: "new"}, "file_path is required"},
			{LcEditFileParams{AppName: "testapp", FilePath: "test.txt", NewString: "new"}, "old_string is required"},
			{LcEditFileParams{AppName: "testapp", FilePath: "test.txt", OldString: "old", NewString: "new", Occurrences: -1}, "occurrences must be non-negative"},
			{LcEditFileParams{AppName: "testapp", FilePath: "nonexistent.txt", OldString: "old", NewString: "new"}, "failed to read file"},
		}
		for _, tt := range tests {
			_, err := LcEditFile(tt.params)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("EditFile(%+v) expected error containing %q, got: %v",
					tt.params, tt.wantErr, err)
			}
		}
	})

	t.Run("string not found", func(t *testing.T) {
		testFile := filepath.Join(appDir, "test5.txt")
		os.WriteFile(testFile, []byte("some content"), 0644)

		params := LcEditFileParams{
			AppName:   "testapp",
			FilePath:  "test5.txt",
			OldString: "nonexistent",
			NewString: "replacement",
		}
		_, err := LcEditFile(params)
		if err == nil || !strings.Contains(err.Error(), "old_string not found") {
			t.Errorf("Expected 'old_string not found' error, got: %v", err)
		}
	})

	t.Run("file size limit", func(t *testing.T) {
		testFile := filepath.Join(appDir, "large.txt")
		os.WriteFile(testFile, []byte(strings.Repeat("a", int(constants.MaxFileSize)+1)), 0644)

		params := LcEditFileParams{
			AppName:   "testapp",
			FilePath:  "large.txt",
			OldString: "a",
			NewString: "b",
		}
		_, err := LcEditFile(params)
		if err == nil || !strings.Contains(err.Error(), "exceeds maximum size") {
			t.Errorf("Expected file size error, got: %v", err)
		}
	})

	t.Run("path traversal attempt", func(t *testing.T) {
		params := LcEditFileParams{
			AppName:   "testapp",
			FilePath:  "../../../etc/passwd",
			OldString: "root",
			NewString: "toor",
		}
		_, err := LcEditFile(params)
		if err == nil || !strings.Contains(err.Error(), "outside app directory") {
			t.Error("Expected error for path traversal attempt")
		}
	})

	t.Run("occurrences greater than actual", func(t *testing.T) {
		testFile := filepath.Join(appDir, "test6.txt")
		os.WriteFile(testFile, []byte("foo bar foo"), 0644)

		params := LcEditFileParams{
			AppName:     "testapp",
			FilePath:    "test6.txt",
			OldString:   "foo",
			NewString:   "baz",
			Occurrences: 10, // More than actual occurrences
		}
		result, err := LcEditFile(params)
		if err != nil {
			t.Fatalf("EditFile() failed: %v", err)
		}
		if result.Replacements != 2 {
			t.Errorf("Replacements = %d; want 2", result.Replacements)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		expected := "baz bar baz"
		if string(content) != expected {
			t.Errorf("File content = %q; want %q", content, expected)
		}
	})
}

// TestLcEditFileCli tests the CLI interface
func TestLcEditFileCli(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "test.txt"), []byte("hello world"), 0644)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	// Save original os.Args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	t.Run("missing arguments", func(t *testing.T) {
		tests := []struct {
			args    []string
			wantErr string
		}{
			{[]string{"cmd", "tool", "edit_file"}, "--app-name is required"},
			{[]string{"cmd", "tool", "edit_file", "--app-name", "testapp"}, "--file-path is required"},
			{[]string{"cmd", "tool", "edit_file", "--app-name", "testapp", "--file-path", "test.txt"}, "--old-string is required"},
			{[]string{"cmd", "tool", "edit_file", "--app-name"}, "--app-name requires a value"},
			{[]string{"cmd", "tool", "edit_file", "--file-path"}, "--file-path requires a value"},
			{[]string{"cmd", "tool", "edit_file", "--old-string"}, "--old-string requires a value"},
			{[]string{"cmd", "tool", "edit_file", "--new-string"}, "--new-string requires a value"},
			{[]string{"cmd", "tool", "edit_file", "--occurrences"}, "--occurrences requires a value"},
			{[]string{"cmd", "tool", "edit_file", "--unknown"}, "unknown option: --unknown"},
		}
		for _, tt := range tests {
			os.Args = tt.args
			err := LcEditFileCli()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("EditFileCli() with args %v expected error containing %q, got: %v",
					tt.args[3:], tt.wantErr, err)
			}
		}
	})

	t.Run("help flag", func(t *testing.T) {
		for _, helpFlag := range []string{"--help", "-h"} {
			os.Args = []string{"cmd", "tool", "edit_file", helpFlag}
			err := LcEditFileCli()
			if err != nil {
				t.Errorf("EditFileCli() with %s should not error, got: %v", helpFlag, err)
			}
		}
	})

	t.Run("successful execution", func(t *testing.T) {
		os.Args = []string{"cmd", "tool", "edit_file", "--app-name", "testapp",
			"--file-path", "test.txt", "--old-string", "hello", "--new-string", "goodbye"}
		err := LcEditFileCli()
		if err != nil {
			t.Errorf("EditFileCli() failed: %v", err)
		}

		// Verify file was edited
		content, _ := os.ReadFile(filepath.Join(appDir, "test.txt"))
		if string(content) != "goodbye world" {
			t.Errorf("File content = %q; want %q", content, "goodbye world")
		}
	})

	t.Run("with occurrences option", func(t *testing.T) {
		testFile := filepath.Join(appDir, "multi.txt")
		os.WriteFile(testFile, []byte("a b a b a"), 0644)

		os.Args = []string{"cmd", "tool", "edit_file", "--app-name", "testapp",
			"--file-path", "multi.txt", "--old-string", "a", "--new-string", "x", "--occurrences", "2"}
		err := LcEditFileCli()
		if err != nil {
			t.Errorf("EditFileCli() failed: %v", err)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		if string(content) != "x b x b a" {
			t.Errorf("File content = %q; want %q", content, "x b x b a")
		}
	})

	t.Run("empty new-string (deletion)", func(t *testing.T) {
		testFile := filepath.Join(appDir, "delete.txt")
		os.WriteFile(testFile, []byte("prefix-content"), 0644)

		os.Args = []string{"cmd", "tool", "edit_file", "--app-name", "testapp",
			"--file-path", "delete.txt", "--old-string", "prefix-", "--new-string", ""}
		err := LcEditFileCli()
		if err != nil {
			t.Errorf("EditFileCli() failed: %v", err)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		if string(content) != "content" {
			t.Errorf("File content = %q; want %q", content, "content")
		}
	})
}

// TestLcEditFileMcp tests the MCP interface wrapper
func TestLcEditFileMcp(t *testing.T) {
	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "edit_file"
	request.Params.Arguments = map[string]any{
		"app_name":   "nonexistent",
		"file_path":  "test.txt",
		"old_string": "old",
		"new_string": "new",
	}

	_, err := LcEditFileMcp(ctx, request)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}
}