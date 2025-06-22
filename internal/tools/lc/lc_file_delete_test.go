package lc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestLcFileDelete tests the core LcFileDelete functionality
func TestLcFileDelete(t *testing.T) {
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

	t.Run("successful delete", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(appDir, "delete-me.txt")
		os.WriteFile(testFile, []byte("test content"), 0644)

		params := LcFileDeleteParams{
			AppName:  "testapp",
			FilePath: "delete-me.txt",
		}

		result, err := LcFileDelete(params)
		if err != nil {
			t.Fatalf("LcDeleteFile() failed: %v", err)
		}

		if !result.Deleted {
			t.Error("Expected Deleted to be true")
		}

		// Verify file doesn't exist
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File still exists after deletion")
		}
	})

	t.Run("delete file in subdirectory", func(t *testing.T) {
		// Create subdirectory and file
		subdir := filepath.Join(appDir, "subdir")
		os.Mkdir(subdir, 0755)
		testFile := filepath.Join(subdir, "file.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		params := LcFileDeleteParams{
			AppName:  "testapp",
			FilePath: "subdir/file.txt",
		}

		result, err := LcFileDelete(params)
		if err != nil {
			t.Fatalf("LcDeleteFile() failed: %v", err)
		}

		if !result.Deleted {
			t.Error("Expected Deleted to be true")
		}

		// Verify file doesn't exist
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File still exists after deletion")
		}
	})

	t.Run("input validation errors", func(t *testing.T) {
		tests := []struct {
			name   string
			params LcFileDeleteParams
		}{
			{"empty app name", LcFileDeleteParams{FilePath: "file.txt"}},
			{"empty file path", LcFileDeleteParams{AppName: "testapp"}},
			{"path traversal", LcFileDeleteParams{AppName: "testapp", FilePath: "../file.txt"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := LcFileDelete(tt.params)
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			})
		}
	})

	t.Run("file not found", func(t *testing.T) {
		params := LcFileDeleteParams{
			AppName:  "testapp",
			FilePath: "nonexistent.txt",
		}

		_, err := LcFileDelete(params)
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("cannot delete directory", func(t *testing.T) {
		// Create directory
		os.Mkdir(filepath.Join(appDir, "testdir"), 0755)

		params := LcFileDeleteParams{
			AppName:  "testapp",
			FilePath: "testdir",
		}

		_, err := LcFileDelete(params)
		if err == nil {
			t.Error("Expected error when trying to delete directory")
		}
	})
}

// TestLcFileDeleteCli tests the CLI interface
func TestLcFileDeleteCli(t *testing.T) {
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
		os.Args = []string{"layered-code", "tool", "lc_delete_file", "--help"}
		err := LcFileDeleteCli()
		if err != nil {
			t.Errorf("Help flag should not return error: %v", err)
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_delete_file", "--app-name", "testapp"}
		err := LcFileDeleteCli()
		if err == nil {
			t.Error("Expected error for missing arguments")
		}
	})

	// Note: Testing with --force to avoid interactive prompt
	t.Run("successful deletion with force", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(appDir, "force-delete.txt")
		os.WriteFile(testFile, []byte("test"), 0644)

		os.Args = []string{"layered-code", "tool", "lc_delete_file", 
			"--app-name", "testapp", 
			"--file-path", "force-delete.txt",
			"--force"}
		
		err := LcFileDeleteCli()
		if err != nil {
			t.Errorf("Delete with force failed: %v", err)
		}

		// Verify file was deleted
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File still exists after deletion")
		}
	})
}

// TestLcFileDeleteMcp tests the MCP interface
func TestLcFileDeleteMcp(t *testing.T) {
	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "lc_delete_file"
	request.Params.Arguments = map[string]any{
		"app_name":  "nonexistent",
		"file_path": "file.txt",
	}

	_, err := LcFileDeleteMcp(ctx, request)
	if err == nil {
		t.Error("Expected error for non-existent app")
	}
}