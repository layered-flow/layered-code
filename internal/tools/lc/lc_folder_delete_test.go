package lc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
)

func TestLcFolderDelete(t *testing.T) {
	// Setup test environment
	originalEnv := os.Getenv(constants.AppsDirectoryEnvVar)
	defer os.Setenv(constants.AppsDirectoryEnvVar, originalEnv)

	// Use a directory within the home directory for testing
	_, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home dir: %v", err)
	}

	tempDir := t.TempDir()
	// Make the temp directory relative to home for testing
	testAppsDir := filepath.Join("test-layered-apps", filepath.Base(tempDir))
	os.Setenv(constants.AppsDirectoryEnvVar, testAppsDir)

	// Ensure apps directory exists
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create test app with folder structure
	testApp := "test-app"
	appPath := filepath.Join(appsDir, testApp)
	if err := os.MkdirAll(appPath, 0755); err != nil {
		t.Fatalf("Failed to create app path: %v", err)
	}

	// Create a folder with some content
	testFolder := "test-folder"
	folderPath := filepath.Join(appPath, testFolder)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		t.Fatalf("Failed to create folder path: %v", err)
	}

	// Add some files to the folder
	file1Path := filepath.Join(folderPath, "file1.txt")
	if err := os.WriteFile(file1Path, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	subFolderPath := filepath.Join(folderPath, "subfolder")
	if err := os.MkdirAll(subFolderPath, 0755); err != nil {
		t.Fatalf("Failed to create subfolder: %v", err)
	}

	file2Path := filepath.Join(subFolderPath, "file2.txt")
	if err := os.WriteFile(file2Path, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	t.Run("successful folder deletion", func(t *testing.T) {
		params := LcFolderDeleteParams{
			AppName:    testApp,
			FolderPath: testFolder,
		}

		result, err := LcFolderDelete(params)
		if err != nil {
			t.Fatalf("Failed to delete folder: %v", err)
		}
		if result.AppName != testApp {
			t.Errorf("Expected AppName %s, got %s", testApp, result.AppName)
		}
		if result.FolderPath != testFolder {
			t.Errorf("Expected FolderPath %s, got %s", testFolder, result.FolderPath)
		}
		if !result.Deleted {
			t.Errorf("Expected Deleted to be true, got false")
		}

		// Verify folder is deleted
		_, err = os.Stat(folderPath)
		if !os.IsNotExist(err) {
			t.Errorf("Expected folder to be deleted, but it still exists")
		}
	})

	t.Run("folder not found", func(t *testing.T) {
		params := LcFolderDeleteParams{
			AppName:    testApp,
			FolderPath: "non-existent-folder",
		}

		_, err := LcFolderDelete(params)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "folder not found") {
			t.Errorf("Expected error to contain 'folder not found', got: %v", err)
		}
	})

	t.Run("empty app name", func(t *testing.T) {
		params := LcFolderDeleteParams{
			AppName:    "",
			FolderPath: testFolder,
		}

		_, err := LcFolderDelete(params)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "app_name is required") {
			t.Errorf("Expected error to contain 'app_name is required', got: %v", err)
		}
	})

	t.Run("empty folder path", func(t *testing.T) {
		params := LcFolderDeleteParams{
			AppName:    testApp,
			FolderPath: "",
		}

		_, err := LcFolderDelete(params)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "folder_path is required") {
			t.Errorf("Expected error to contain 'folder_path is required', got: %v", err)
		}
	})

	t.Run("directory traversal attempt", func(t *testing.T) {
		params := LcFolderDeleteParams{
			AppName:    testApp,
			FolderPath: "../other-app/folder",
		}

		_, err := LcFolderDelete(params)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "directory traversal is not allowed") {
			t.Errorf("Expected error to contain 'directory traversal is not allowed', got: %v", err)
		}
	})

	t.Run("attempt to delete file instead of folder", func(t *testing.T) {
		// Create a test file
		testFile := "test-file.txt"
		filePath := filepath.Join(appPath, testFile)
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		params := LcFolderDeleteParams{
			AppName:    testApp,
			FolderPath: testFile,
		}

		_, err := LcFolderDelete(params)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "path is not a directory") {
			t.Errorf("Expected error to contain 'path is not a directory', got: %v", err)
		}
	})

	t.Run("attempt to delete app root directory", func(t *testing.T) {
		params := LcFolderDeleteParams{
			AppName:    testApp,
			FolderPath: ".",
		}

		_, err := LcFolderDelete(params)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !strings.Contains(err.Error(), "cannot delete the app root directory") {
			t.Errorf("Expected error to contain 'cannot delete the app root directory', got: %v", err)
		}
	})
}

