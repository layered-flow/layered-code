package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitShow(t *testing.T) {
	// Setup test environment
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create a test app directory
	testApp := "test-git-show-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath) // Cleanup

	// Create app directory
	if err := os.MkdirAll(testAppPath, 0755); err != nil {
		t.Fatalf("Failed to create test app directory: %v", err)
	}

	// Test 1: Non-git repository
	t.Run("NonGitRepo", func(t *testing.T) {
		result, err := GitShow(testApp, "HEAD")
		if err != nil {
			t.Fatalf("GitShow failed: %v", err)
		}

		if result.IsRepo {
			t.Error("Expected IsRepo to be false for non-git directory")
		}

		if result.Success {
			t.Error("Expected Success to be false for non-git directory")
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

	// Test 2: Empty repository (no commits)
	t.Run("EmptyRepo", func(t *testing.T) {
		result, err := GitShow(testApp, "HEAD")
		if err != nil {
			t.Fatalf("GitShow failed: %v", err)
		}

		if !result.IsRepo {
			t.Error("Expected IsRepo to be true for git directory")
		}

		if result.Success {
			t.Error("Expected Success to be false for empty repository")
		}

		if !strings.Contains(result.Message, "Failed to show commit") {
			t.Errorf("Expected 'Failed to show commit' message, got: %s", result.Message)
		}
	})

	// Create and commit a test file
	testFile := filepath.Join(testAppPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content line 1\ntest content line 2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	addCmd := exec.Command("git", "add", "test.txt")
	addCmd.Dir = testAppPath
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit with test file")
	commitCmd.Dir = testAppPath
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	// Test 3: Show HEAD commit
	t.Run("ShowHEAD", func(t *testing.T) {
		result, err := GitShow(testApp, "HEAD")
		if err != nil {
			t.Fatalf("GitShow failed: %v", err)
		}

		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}

		if !result.Success {
			t.Error("Expected Success to be true")
		}

		if result.Content == "" {
			t.Error("Expected Content to be non-empty")
		}

		if result.Hash == "" {
			t.Error("Expected Hash to be non-empty")
		}

		if result.Author == "" {
			t.Error("Expected Author to be non-empty")
		}

		if result.Date == "" {
			t.Error("Expected Date to be non-empty")
		}

		if result.Subject != "Initial commit with test file" {
			t.Errorf("Expected Subject to be 'Initial commit with test file', got: %s", result.Subject)
		}

		if result.CommitRef != "HEAD" {
			t.Errorf("Expected CommitRef to be 'HEAD', got: %s", result.CommitRef)
		}

		// Check that content contains expected parts
		if !strings.Contains(result.Content, "Initial commit with test file") {
			t.Error("Expected Content to contain commit message")
		}

		if !strings.Contains(result.Content, "test content line 1") {
			t.Error("Expected Content to contain file diff")
		}
	})

	// Test 4: Show with default (empty) commit ref
	t.Run("ShowDefault", func(t *testing.T) {
		result, err := GitShow(testApp, "")
		if err != nil {
			t.Fatalf("GitShow failed: %v", err)
		}

		if !result.Success {
			t.Error("Expected Success to be true for default commit ref")
		}

		if result.CommitRef != "HEAD" {
			t.Errorf("Expected CommitRef to be 'HEAD', got: %s", result.CommitRef)
		}
	})

	// Create another commit for more testing
	if err := os.WriteFile(testFile, []byte("modified content line 1\nmodified content line 2\nnew line 3"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	addCmd = exec.Command("git", "add", "test.txt")
	addCmd.Dir = testAppPath
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add modified file: %v", err)
	}

	commitCmd = exec.Command("git", "commit", "-m", "Second commit with modifications")
	commitCmd.Dir = testAppPath
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to commit modified file: %v", err)
	}

	// Test 5: Show specific commit by partial hash
	t.Run("ShowByPartialHash", func(t *testing.T) {
		// Get the first commit hash
		logCmd := exec.Command("git", "log", "--oneline", "--reverse")
		logCmd.Dir = testAppPath
		output, err := logCmd.Output()
		if err != nil {
			t.Fatalf("Failed to get git log: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) < 1 {
			t.Fatal("Expected at least one commit in log")
		}

		firstCommitHash := strings.Fields(lines[0])[0]
		shortHash := firstCommitHash[:7] // Use short hash

		result, err := GitShow(testApp, shortHash)
		if err != nil {
			t.Fatalf("GitShow failed: %v", err)
		}

		if !result.Success {
			t.Error("Expected Success to be true for valid commit hash")
		}

		if result.Subject != "Initial commit with test file" {
			t.Errorf("Expected Subject to be 'Initial commit with test file', got: %s", result.Subject)
		}

		if result.CommitRef != shortHash {
			t.Errorf("Expected CommitRef to be '%s', got: %s", shortHash, result.CommitRef)
		}
	})

	// Test 6: Show invalid commit
	t.Run("ShowInvalidCommit", func(t *testing.T) {
		result, err := GitShow(testApp, "invalid-commit-hash")
		if err != nil {
			t.Fatalf("GitShow failed: %v", err)
		}

		if result.Success {
			t.Error("Expected Success to be false for invalid commit")
		}

		if !result.IsRepo {
			t.Error("Expected IsRepo to be true even for invalid commit")
		}

		if !strings.Contains(result.Message, "Failed to show commit") {
			t.Errorf("Expected error message about failed commit, got: %s", result.Message)
		}

		if result.CommitRef != "invalid-commit-hash" {
			t.Errorf("Expected CommitRef to be 'invalid-commit-hash', got: %s", result.CommitRef)
		}
	})

	// Test 7: Invalid app name
	t.Run("InvalidAppName", func(t *testing.T) {
		_, err := GitShow("../invalid-app", "HEAD")
		if err == nil {
			t.Error("Expected error for invalid app name")
		}
	})
}

func TestParseCommitInfo(t *testing.T) {
	// Test parsing commit info from git show output
	testOutput := `commit 1234567890abcdef1234567890abcdef12345678
Author:     Test User <test@example.com>
AuthorDate: Mon Jan 1 12:00:00 2024 +0000
Commit:     Test User <test@example.com>
CommitDate: Mon Jan 1 12:00:00 2024 +0000

    Initial commit with test file

diff --git a/test.txt b/test.txt
new file mode 100644
index 0000000..abcdef1
--- /dev/null
+++ b/test.txt
@@ -0,0 +1,2 @@
+test content line 1
+test content line 2`

	hash, author, date, subject := parseCommitInfo(testOutput)

	expectedHash := "1234567890abcdef1234567890abcdef12345678"
	if hash != expectedHash {
		t.Errorf("Expected hash '%s', got '%s'", expectedHash, hash)
	}

	expectedAuthor := "Test User <test@example.com>"
	if author != expectedAuthor {
		t.Errorf("Expected author '%s', got '%s'", expectedAuthor, author)
	}

	expectedDate := "Mon Jan 1 12:00:00 2024 +0000"
	if date != expectedDate {
		t.Errorf("Expected date '%s', got '%s'", expectedDate, date)
	}

	expectedSubject := "Initial commit with test file"
	if subject != expectedSubject {
		t.Errorf("Expected subject '%s', got '%s'", expectedSubject, subject)
	}
}

func TestParseCommitInfoEmptySubject(t *testing.T) {
	// Test parsing commit info with no commit message
	testOutput := `commit 1234567890abcdef1234567890abcdef12345678
Author:     Test User <test@example.com>
AuthorDate: Mon Jan 1 12:00:00 2024 +0000
Commit:     Test User <test@example.com>
CommitDate: Mon Jan 1 12:00:00 2024 +0000

diff --git a/test.txt b/test.txt`

	hash, author, date, subject := parseCommitInfo(testOutput)

	if hash != "1234567890abcdef1234567890abcdef12345678" {
		t.Errorf("Expected hash to be parsed correctly, got '%s'", hash)
	}

	if author != "Test User <test@example.com>" {
		t.Errorf("Expected author to be parsed correctly, got '%s'", author)
	}

	if date != "Mon Jan 1 12:00:00 2024 +0000" {
		t.Errorf("Expected date to be parsed correctly, got '%s'", date)
	}

	if subject != "" {
		t.Errorf("Expected empty subject, got '%s'", subject)
	}
}