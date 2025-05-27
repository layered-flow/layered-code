package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitDiff(t *testing.T) {
	// Setup test environment
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create a test app directory
	testApp := "test-git-diff-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath) // Cleanup

	// Create app directory
	if err := os.MkdirAll(testAppPath, 0755); err != nil {
		t.Fatalf("Failed to create test app directory: %v", err)
	}

	// Test 1: Non-git repository
	t.Run("NonGitRepo", func(t *testing.T) {
		result, err := GitDiff(testApp, false, "")
		if err != nil {
			t.Fatalf("GitDiff failed: %v", err)
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

	// Test 2: Empty repository (no diff)
	t.Run("EmptyRepo", func(t *testing.T) {
		result, err := GitDiff(testApp, false, "")
		if err != nil {
			t.Fatalf("GitDiff failed: %v", err)
		}

		if !result.IsRepo {
			t.Error("Expected IsRepo to be true for git directory")
		}

		if result.HasDiff {
			t.Error("Expected no diff in empty repository")
		}
	})

	// Create and commit a file
	testFile := filepath.Join(testAppPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial content\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = testAppPath
	addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = testAppPath
	commitCmd.Run()

	// Test 3: Working directory diff
	t.Run("WorkingDirDiff", func(t *testing.T) {
		// Modify the file
		if err := os.WriteFile(testFile, []byte("modified content\n"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		result, err := GitDiff(testApp, false, "")
		if err != nil {
			t.Fatalf("GitDiff failed: %v", err)
		}

		if !result.HasDiff {
			t.Error("Expected diff for modified file")
		}

		if !strings.Contains(result.Diff, "test.txt") {
			t.Error("Expected diff to contain test.txt")
		}

		if !strings.Contains(result.Diff, "-initial content") {
			t.Error("Expected diff to show removed line")
		}

		if !strings.Contains(result.Diff, "+modified content") {
			t.Error("Expected diff to show added line")
		}

		if result.FileCount != 1 {
			t.Errorf("Expected file count to be 1, got %d", result.FileCount)
		}
	})

	// Test 4: Staged diff
	t.Run("StagedDiff", func(t *testing.T) {
		// Stage the modified file
		addCmd := exec.Command("git", "add", "test.txt")
		addCmd.Dir = testAppPath
		addCmd.Run()

		result, err := GitDiff(testApp, true, "")
		if err != nil {
			t.Fatalf("GitDiff failed: %v", err)
		}

		if !result.HasDiff {
			t.Error("Expected diff for staged file")
		}

		if !strings.Contains(result.Diff, "test.txt") {
			t.Error("Expected diff to contain test.txt")
		}
	})

	// Test 5: Specific file diff
	t.Run("SpecificFileDiff", func(t *testing.T) {
		// Reset the working directory by committing staged changes
		commitCmd := exec.Command("git", "commit", "-m", "Commit for test")
		commitCmd.Dir = testAppPath
		commitCmd.Run()

		// Create and add another file
		testFile2 := filepath.Join(testAppPath, "test2.txt")
		if err := os.WriteFile(testFile2, []byte("another file\n"), 0644); err != nil {
			t.Fatalf("Failed to create test file 2: %v", err)
		}

		// Add and commit test2.txt
		addCmd := exec.Command("git", "add", "test2.txt")
		addCmd.Dir = testAppPath
		addCmd.Run()

		commitCmd = exec.Command("git", "commit", "-m", "Add test2.txt")
		commitCmd.Dir = testAppPath
		commitCmd.Run()

		// Now modify test2.txt to create a diff
		if err := os.WriteFile(testFile2, []byte("modified another file\n"), 0644); err != nil {
			t.Fatalf("Failed to modify test file 2: %v", err)
		}

		result, err := GitDiff(testApp, false, "test2.txt")
		if err != nil {
			t.Fatalf("GitDiff failed: %v", err)
		}

		if !result.HasDiff {
			t.Error("Expected diff for specific file")
		}

		if !strings.Contains(result.Diff, "test2.txt") {
			t.Error("Expected diff to contain test2.txt")
		}

		if strings.Contains(result.Diff, "test.txt") {
			t.Error("Expected diff to NOT contain test.txt")
		}
	})

	// Test 6: Invalid file path
	t.Run("InvalidFilePath", func(t *testing.T) {
		_, err := GitDiff(testApp, false, "../outside")
		if err == nil {
			t.Error("Expected error for path outside app directory")
		}
	})
}