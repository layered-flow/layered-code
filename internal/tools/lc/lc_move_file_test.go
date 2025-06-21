package lc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestLcMoveFile tests the core LcMoveFile functionality
func TestLcMoveFile(t *testing.T) {
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

	t.Run("successful rename", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(appDir, "old-name.txt")
		os.WriteFile(testFile, []byte("test content"), 0644)

		params := LcMoveFileParams{
			AppName:    "testapp",
			SourcePath: "old-name.txt",
			DestPath:   "new-name.txt",
		}

		result, err := LcMoveFile(params)
		if err != nil {
			t.Fatalf("LcMoveFile() failed: %v", err)
		}

		if !result.IsRename {
			t.Error("Expected IsRename to be true for same directory move")
		}

		// Verify old file doesn't exist
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("Old file still exists")
		}

		// Verify new file exists
		newFile := filepath.Join(appDir, "new-name.txt")
		if _, err := os.Stat(newFile); err != nil {
			t.Error("New file doesn't exist")
		}

		// Verify content
		content, _ := os.ReadFile(newFile)
		if string(content) != "test content" {
			t.Error("File content changed during move")
		}
	})

	t.Run("successful move to subdirectory", func(t *testing.T) {
		// Create test file and directory
		testFile := filepath.Join(appDir, "file.txt")
		os.WriteFile(testFile, []byte("test content"), 0644)
		os.Mkdir(filepath.Join(appDir, "subdir"), 0755)

		params := LcMoveFileParams{
			AppName:    "testapp",
			SourcePath: "file.txt",
			DestPath:   "subdir/file.txt",
		}

		result, err := LcMoveFile(params)
		if err != nil {
			t.Fatalf("LcMoveFile() failed: %v", err)
		}

		if result.IsRename {
			t.Error("Expected IsRename to be false for different directory move")
		}

		// Verify file moved
		newFile := filepath.Join(appDir, "subdir/file.txt")
		if _, err := os.Stat(newFile); err != nil {
			t.Error("File not moved to subdirectory")
		}
	})

	t.Run("input validation errors", func(t *testing.T) {
		tests := []struct {
			name   string
			params LcMoveFileParams
		}{
			{"empty app name", LcMoveFileParams{SourcePath: "file.txt", DestPath: "new.txt"}},
			{"empty source", LcMoveFileParams{AppName: "testapp", DestPath: "new.txt"}},
			{"empty dest", LcMoveFileParams{AppName: "testapp", SourcePath: "file.txt"}},
			{"path traversal in source", LcMoveFileParams{AppName: "testapp", SourcePath: "../file.txt", DestPath: "new.txt"}},
			{"path traversal in dest", LcMoveFileParams{AppName: "testapp", SourcePath: "file.txt", DestPath: "../new.txt"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := LcMoveFile(tt.params)
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			})
		}
	})

	t.Run("source file not found", func(t *testing.T) {
		params := LcMoveFileParams{
			AppName:    "testapp",
			SourcePath: "nonexistent.txt",
			DestPath:   "new.txt",
		}

		_, err := LcMoveFile(params)
		if err == nil {
			t.Error("Expected error for nonexistent source file")
		}
	})

	t.Run("destination already exists", func(t *testing.T) {
		// Create both files
		os.WriteFile(filepath.Join(appDir, "source.txt"), []byte("source"), 0644)
		os.WriteFile(filepath.Join(appDir, "dest.txt"), []byte("dest"), 0644)

		params := LcMoveFileParams{
			AppName:    "testapp",
			SourcePath: "source.txt",
			DestPath:   "dest.txt",
		}

		_, err := LcMoveFile(params)
		if err == nil {
			t.Error("Expected error when destination exists")
		}
	})

	t.Run("cannot move directory", func(t *testing.T) {
		// Create directory
		os.Mkdir(filepath.Join(appDir, "testdir"), 0755)

		params := LcMoveFileParams{
			AppName:    "testapp",
			SourcePath: "testdir",
			DestPath:   "newdir",
		}

		_, err := LcMoveFile(params)
		if err == nil {
			t.Error("Expected error when trying to move directory")
		}
	})

	t.Run("successful overwrite", func(t *testing.T) {
		// Create source and destination files
		os.WriteFile(filepath.Join(appDir, "source.txt"), []byte("source content"), 0644)
		os.WriteFile(filepath.Join(appDir, "existing.txt"), []byte("existing content"), 0644)

		params := LcMoveFileParams{
			AppName:    "testapp",
			SourcePath: "source.txt",
			DestPath:   "existing.txt",
			Overwrite:  true,
		}

		_, err := LcMoveFile(params)
		if err != nil {
			t.Fatalf("LcMoveFile() with overwrite failed: %v", err)
		}

		// Verify source doesn't exist
		if _, err := os.Stat(filepath.Join(appDir, "source.txt")); !os.IsNotExist(err) {
			t.Error("Source file still exists after move")
		}

		// Verify destination has source content
		content, _ := os.ReadFile(filepath.Join(appDir, "existing.txt"))
		if string(content) != "source content" {
			t.Error("Destination file content not updated after overwrite")
		}
	})

	t.Run("overwrite flag false still fails", func(t *testing.T) {
		// Create both files
		os.WriteFile(filepath.Join(appDir, "source2.txt"), []byte("source"), 0644)
		os.WriteFile(filepath.Join(appDir, "dest2.txt"), []byte("dest"), 0644)

		params := LcMoveFileParams{
			AppName:    "testapp",
			SourcePath: "source2.txt",
			DestPath:   "dest2.txt",
			Overwrite:  false,
		}

		_, err := LcMoveFile(params)
		if err == nil {
			t.Error("Expected error when destination exists and overwrite is false")
		}
	})
}

// TestLcMoveFileCli tests the CLI interface
func TestLcMoveFileCli(t *testing.T) {
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
		os.Args = []string{"layered-code", "tool", "lc_move_file", "--help"}
		err := LcMoveFileCli()
		if err != nil {
			t.Errorf("Help flag should not return error: %v", err)
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_move_file", "--app-name", "testapp"}
		err := LcMoveFileCli()
		if err == nil {
			t.Error("Expected error for missing arguments")
		}
	})
}

// TestLcMoveFileMcp tests the MCP interface
func TestLcMoveFileMcp(t *testing.T) {
	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "lc_move_file"
	request.Params.Arguments = map[string]any{
		"app_name":    "nonexistent",
		"source_path": "file.txt",
		"dest_path":   "new.txt",
	}

	_, err := LcMoveFileMcp(ctx, request)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}
}