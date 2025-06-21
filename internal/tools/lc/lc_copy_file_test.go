package lc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestLcCopyFile tests the core LcCopyFile functionality
func TestLcCopyFile(t *testing.T) {
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

	t.Run("successful copy", func(t *testing.T) {
		// Create test file
		sourceFile := filepath.Join(appDir, "source.txt")
		os.WriteFile(sourceFile, []byte("test content"), 0644)

		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "source.txt",
			DestPath:   "copy.txt",
		}

		result, err := LcCopyFile(params)
		if err != nil {
			t.Fatalf("LcCopyFile() failed: %v", err)
		}

		if result.BytesCopied != 12 {
			t.Errorf("Expected 12 bytes copied, got %d", result.BytesCopied)
		}

		// Verify both files exist
		if _, err := os.Stat(sourceFile); err != nil {
			t.Error("Source file doesn't exist")
		}

		destFile := filepath.Join(appDir, "copy.txt")
		if _, err := os.Stat(destFile); err != nil {
			t.Error("Destination file doesn't exist")
		}

		// Verify content
		content, _ := os.ReadFile(destFile)
		if string(content) != "test content" {
			t.Error("Copied file has wrong content")
		}
	})

	t.Run("copy to subdirectory", func(t *testing.T) {
		// Create test file
		sourceFile := filepath.Join(appDir, "file.txt")
		os.WriteFile(sourceFile, []byte("test"), 0644)

		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "file.txt",
			DestPath:   "subdir/file.txt",
		}

		_, err := LcCopyFile(params)
		if err != nil {
			t.Fatalf("LcCopyFile() failed: %v", err)
		}

		// Verify file copied
		destFile := filepath.Join(appDir, "subdir/file.txt")
		if _, err := os.Stat(destFile); err != nil {
			t.Error("File not copied to subdirectory")
		}

		// Verify source still exists
		if _, err := os.Stat(sourceFile); err != nil {
			t.Error("Source file was removed (should still exist)")
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		// Create both files
		sourceFile := filepath.Join(appDir, "source2.txt")
		destFile := filepath.Join(appDir, "dest2.txt")
		os.WriteFile(sourceFile, []byte("new content"), 0644)
		os.WriteFile(destFile, []byte("old content"), 0644)

		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "source2.txt",
			DestPath:   "dest2.txt",
			Overwrite:  true,
		}

		_, err := LcCopyFile(params)
		if err != nil {
			t.Fatalf("LcCopyFile() with overwrite failed: %v", err)
		}

		// Verify content was overwritten
		content, _ := os.ReadFile(destFile)
		if string(content) != "new content" {
			t.Error("File was not overwritten")
		}
	})

	t.Run("input validation errors", func(t *testing.T) {
		tests := []struct {
			name   string
			params LcCopyFileParams
		}{
			{"empty app name", LcCopyFileParams{SourcePath: "file.txt", DestPath: "copy.txt"}},
			{"empty source", LcCopyFileParams{AppName: "testapp", DestPath: "copy.txt"}},
			{"empty dest", LcCopyFileParams{AppName: "testapp", SourcePath: "file.txt"}},
			{"path traversal in source", LcCopyFileParams{AppName: "testapp", SourcePath: "../file.txt", DestPath: "copy.txt"}},
			{"path traversal in dest", LcCopyFileParams{AppName: "testapp", SourcePath: "file.txt", DestPath: "../copy.txt"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := LcCopyFile(tt.params)
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			})
		}
	})

	t.Run("source file not found", func(t *testing.T) {
		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "nonexistent.txt",
			DestPath:   "copy.txt",
		}

		_, err := LcCopyFile(params)
		if err == nil {
			t.Error("Expected error for nonexistent source file")
		}
	})

	t.Run("destination exists without overwrite", func(t *testing.T) {
		// Create both files
		os.WriteFile(filepath.Join(appDir, "src.txt"), []byte("source"), 0644)
		os.WriteFile(filepath.Join(appDir, "dst.txt"), []byte("dest"), 0644)

		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "src.txt",
			DestPath:   "dst.txt",
			Overwrite:  false,
		}

		_, err := LcCopyFile(params)
		if err == nil {
			t.Error("Expected error when destination exists without overwrite")
		}
	})

	t.Run("cannot copy directory", func(t *testing.T) {
		// Create directory
		os.Mkdir(filepath.Join(appDir, "testdir2"), 0755)

		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "testdir2",
			DestPath:   "copydir",
		}

		_, err := LcCopyFile(params)
		if err == nil {
			t.Error("Expected error when trying to copy directory")
		}
	})

	t.Run("cannot copy file to itself", func(t *testing.T) {
		// Create file
		os.WriteFile(filepath.Join(appDir, "same.txt"), []byte("test"), 0644)

		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "same.txt",
			DestPath:   "same.txt",
		}

		_, err := LcCopyFile(params)
		if err == nil {
			t.Error("Expected error when copying file to itself")
		}
	})

	t.Run("file size limit", func(t *testing.T) {
		// Create a large file (over 10MB)
		largeFile := filepath.Join(appDir, "large.txt")
		data := strings.Repeat("x", 10*1024*1024+1) // 10MB + 1 byte
		os.WriteFile(largeFile, []byte(data), 0644)

		params := LcCopyFileParams{
			AppName:    "testapp",
			SourcePath: "large.txt",
			DestPath:   "large-copy.txt",
		}

		_, err := LcCopyFile(params)
		if err == nil {
			t.Error("Expected error for file exceeding size limit")
		}
	})
}

// TestLcCopyFileCli tests the CLI interface
func TestLcCopyFileCli(t *testing.T) {
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

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("help flag", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_copy_file", "--help"}
		err := LcCopyFileCli()
		if err != nil {
			t.Errorf("Help flag should not return error: %v", err)
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_copy_file", "--app-name", "testapp"}
		err := LcCopyFileCli()
		if err == nil {
			t.Error("Expected error for missing arguments")
		}
	})

	t.Run("successful copy", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(appDir, "cli-source.txt")
		os.WriteFile(testFile, []byte("cli test"), 0644)

		os.Args = []string{"layered-code", "tool", "lc_copy_file",
			"--app-name", "testapp",
			"--source", "cli-source.txt",
			"--dest", "cli-dest.txt"}

		err := LcCopyFileCli()
		if err != nil {
			t.Errorf("Copy failed: %v", err)
		}

		// Verify file was copied
		destFile := filepath.Join(appDir, "cli-dest.txt")
		if _, err := os.Stat(destFile); err != nil {
			t.Error("Destination file not created")
		}
	})
}

// TestLcCopyFileMcp tests the MCP interface
func TestLcCopyFileMcp(t *testing.T) {
	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "lc_copy_file"
	request.Params.Arguments = map[string]any{
		"app_name":    "nonexistent",
		"source_path": "file.txt",
		"dest_path":   "copy.txt",
	}

	_, err := LcCopyFileMcp(ctx, request)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}
}