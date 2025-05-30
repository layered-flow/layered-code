package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitLog(t *testing.T) {
	// Setup test environment
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create a test app directory
	testApp := "test-git-log-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath) // Cleanup

	// Create app directory
	if err := os.MkdirAll(testAppPath, 0755); err != nil {
		t.Fatalf("Failed to create test app directory: %v", err)
	}

	// Test 1: Non-git repository
	t.Run("NonGitRepo", func(t *testing.T) {
		result, err := GitLog(testApp, 10, false)
		if err != nil {
			t.Fatalf("GitLog failed: %v", err)
		}

		if result.IsRepo {
			t.Error("Expected IsRepo to be false for non-git directory")
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

	// Test 2: Empty repository (no commits)
	t.Run("EmptyRepo", func(t *testing.T) {
		result, err := GitLog(testApp, 10, false)
		if err != nil {
			t.Fatalf("GitLog failed: %v", err)
		}

		if !result.IsRepo {
			t.Error("Expected IsRepo to be true for git directory")
		}

		if len(result.Commits) != 0 {
			t.Errorf("Expected 0 commits, got %d", len(result.Commits))
		}

		if result.Message != "No commits yet" {
			t.Errorf("Expected 'No commits yet' message, got: %s", result.Message)
		}
	})

	// Create some commits
	for i := 1; i <= 5; i++ {
		testFile := filepath.Join(testAppPath, "test.txt")
		content := []byte(fmt.Sprintf("Content %d\n", i))
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		addCmd := exec.Command("git", "add", "test.txt")
		addCmd.Dir = testAppPath
		addCmd.Run()

		commitCmd := exec.Command("git", "commit", "-m", fmt.Sprintf("Commit %d", i))
		commitCmd.Dir = testAppPath
		commitCmd.Run()
	}

	// Test 3: Log with limit
	t.Run("LogWithLimit", func(t *testing.T) {
		result, err := GitLog(testApp, 3, false)
		if err != nil {
			t.Fatalf("GitLog failed: %v", err)
		}

		if len(result.Commits) != 3 {
			t.Errorf("Expected 3 commits with limit, got %d", len(result.Commits))
		}

		// Check that we have all the expected fields
		for _, commit := range result.Commits {
			if commit.Hash == "" {
				t.Error("Expected non-empty commit hash")
			}
			if commit.Author == "" {
				t.Error("Expected non-empty author")
			}
			if commit.Date == "" {
				t.Error("Expected non-empty date")
			}
			if commit.Message == "" {
				t.Error("Expected non-empty message")
			}
		}
	})

	// Test 4: Oneline format
	t.Run("OnelineFormat", func(t *testing.T) {
		result, err := GitLog(testApp, 2, true)
		if err != nil {
			t.Fatalf("GitLog failed: %v", err)
		}

		if len(result.Commits) != 2 {
			t.Errorf("Expected 2 commits, got %d", len(result.Commits))
		}

		// In oneline format, we should only have hash and message
		for _, commit := range result.Commits {
			if commit.Hash == "" {
				t.Error("Expected non-empty commit hash")
			}
			if commit.Message == "" {
				t.Error("Expected non-empty message")
			}
			// Author and Date should be empty in oneline format
			if commit.Author != "" {
				t.Error("Expected empty author in oneline format")
			}
			if commit.Date != "" {
				t.Error("Expected empty date in oneline format")
			}
		}
	})
}