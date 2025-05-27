package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitCommit(t *testing.T) {
	// Setup test environment
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create a test app directory
	testApp := "test-git-commit-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath) // Cleanup

	// Create app directory
	if err := os.MkdirAll(testAppPath, 0755); err != nil {
		t.Fatalf("Failed to create test app directory: %v", err)
	}

	// Test 1: Non-git repository
	t.Run("NonGitRepo", func(t *testing.T) {
		result, err := GitCommit(testApp, "Test message", false)
		if err != nil {
			t.Fatalf("GitCommit failed: %v", err)
		}

		if result.IsRepo {
			t.Error("Expected IsRepo to be false for non-git directory")
		}

		if result.Success {
			t.Error("Expected Success to be false for non-git directory")
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

	// Test 2: No staged changes
	t.Run("NoStagedChanges", func(t *testing.T) {
		result, err := GitCommit(testApp, "Test message", false)
		if err != nil {
			t.Fatalf("GitCommit failed: %v", err)
		}

		if !result.IsRepo {
			t.Error("Expected IsRepo to be true for git directory")
		}

		if result.Success {
			t.Error("Expected Success to be false when no staged changes")
		}

		if !strings.Contains(result.Message, "No staged changes") {
			t.Errorf("Expected message about no staged changes, got: %s", result.Message)
		}
	})

	// Test 3: Missing message
	t.Run("MissingMessage", func(t *testing.T) {
		_, err := GitCommit(testApp, "", false)
		if err == nil {
			t.Error("Expected error for missing commit message")
		}
	})

	// Create and stage a file
	testFile := filepath.Join(testAppPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = testAppPath
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	// Test 4: Successful commit
	t.Run("SuccessfulCommit", func(t *testing.T) {
		result, err := GitCommit(testApp, "Initial commit", false)
		if err != nil {
			t.Fatalf("GitCommit failed: %v", err)
		}

		if !result.Success {
			t.Errorf("Expected Success to be true, error: %s", result.Error)
		}

		if result.CommitHash == "" {
			t.Error("Expected non-empty commit hash")
		}

		if len(result.CommitHash) != 7 {
			t.Errorf("Expected short hash (7 chars), got: %s", result.CommitHash)
		}
	})

	// Test 5: Amend commit
	t.Run("AmendCommit", func(t *testing.T) {
		// Get original commit hash
		hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		hashCmd.Dir = testAppPath
		origHashOutput, _ := hashCmd.Output()
		origHash := strings.TrimSpace(string(origHashOutput))

		// Modify and stage file
		if err := os.WriteFile(testFile, []byte("modified content\n"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		addCmd := exec.Command("git", "add", "test.txt")
		addCmd.Dir = testAppPath
		addCmd.Run()

		// Amend the commit
		result, err := GitCommit(testApp, "Amended commit", true)
		if err != nil {
			t.Fatalf("GitCommit failed: %v", err)
		}

		if !result.Success {
			t.Error("Expected Success to be true for amend")
		}

		// Check that hash changed
		if result.CommitHash == origHash {
			t.Error("Expected different commit hash after amend")
		}
	})

	// Test 6: Amend without message (--no-edit)
	t.Run("AmendNoEdit", func(t *testing.T) {
		// Modify and stage file again
		if err := os.WriteFile(testFile, []byte("another change\n"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		addCmd := exec.Command("git", "add", "test.txt")
		addCmd.Dir = testAppPath
		addCmd.Run()

		// Amend without changing message
		result, err := GitCommit(testApp, "", true)
		if err != nil {
			t.Fatalf("GitCommit failed: %v", err)
		}

		if !result.Success {
			t.Error("Expected Success to be true for amend with no message")
		}
	})
}