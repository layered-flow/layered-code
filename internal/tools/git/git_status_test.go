package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitStatus(t *testing.T) {
	// Setup test environment
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create a test app directory
	testApp := "test-git-status-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath) // Cleanup

	// Create app directory
	if err := os.MkdirAll(testAppPath, 0755); err != nil {
		t.Fatalf("Failed to create test app directory: %v", err)
	}

	// Test 1: Non-git repository
	t.Run("NonGitRepo", func(t *testing.T) {
		result, err := GitStatus(testApp)
		if err != nil {
			t.Fatalf("GitStatus failed: %v", err)
		}

		if result.IsRepo {
			t.Error("Expected IsRepo to be false for non-git directory")
		}

		if result.Message == "" {
			t.Error("Expected message for non-git repository")
		}
	})

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = testAppPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	// Configure git user for the test repo
	configCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configCmd.Dir = testAppPath
	configCmd.Run()

	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = testAppPath
	configCmd.Run()

	// Test 2: Clean repository
	t.Run("CleanRepo", func(t *testing.T) {
		result, err := GitStatus(testApp)
		if err != nil {
			t.Fatalf("GitStatus failed: %v", err)
		}

		if !result.IsRepo {
			t.Error("Expected IsRepo to be true for git directory")
		}

		if len(result.Staged) != 0 || len(result.Modified) != 0 || len(result.Untracked) != 0 {
			t.Error("Expected empty arrays for clean repository")
		}
	})

	// Test 3: Repository with untracked file
	t.Run("UntrackedFile", func(t *testing.T) {
		// Create an untracked file
		testFile := filepath.Join(testAppPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := GitStatus(testApp)
		if err != nil {
			t.Fatalf("GitStatus failed: %v", err)
		}

		if len(result.Untracked) != 1 || result.Untracked[0] != "test.txt" {
			t.Errorf("Expected untracked file 'test.txt', got: %v", result.Untracked)
		}
	})

	// Test 4: Repository with staged file
	t.Run("StagedFile", func(t *testing.T) {
		// Stage the file
		addCmd := exec.Command("git", "add", "test.txt")
		addCmd.Dir = testAppPath
		if err := addCmd.Run(); err != nil {
			t.Fatalf("Failed to stage file: %v", err)
		}

		result, err := GitStatus(testApp)
		if err != nil {
			t.Fatalf("GitStatus failed: %v", err)
		}

		if len(result.Staged) != 1 || result.Staged[0] != "test.txt" {
			t.Errorf("Expected staged file 'test.txt', got: %v", result.Staged)
		}
	})

	// Test 5: Repository with committed and modified file
	t.Run("ModifiedFile", func(t *testing.T) {
		// Commit the file
		commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
		commitCmd.Dir = testAppPath
		if err := commitCmd.Run(); err != nil {
			t.Fatalf("Failed to commit file: %v", err)
		}

		// Modify the file
		testFile := filepath.Join(testAppPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("modified content"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		result, err := GitStatus(testApp)
		if err != nil {
			t.Fatalf("GitStatus failed: %v", err)
		}

		if len(result.Modified) != 1 || result.Modified[0] != "test.txt" {
			t.Errorf("Expected modified file 'test.txt', got: %v", result.Modified)
		}
	})

	// Test 6: Invalid app name
	t.Run("InvalidAppName", func(t *testing.T) {
		_, err := GitStatus("../invalid-app")
		if err == nil {
			t.Error("Expected error for invalid app name")
		}
	})
}