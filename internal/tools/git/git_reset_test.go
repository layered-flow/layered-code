package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

// Test helper functions
func setupTestRepo(t *testing.T) string {
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("failed to ensure apps directory: %v", err)
	}
	
	testApp := fmt.Sprintf("test-git-%d", os.Getpid())
	testAppPath := filepath.Join(appsDir, testApp)
	
	// Clean up any existing directory
	os.RemoveAll(testAppPath)
	
	// Create directory and init git
	if err := os.MkdirAll(testAppPath, 0755); err != nil {
		t.Fatalf("failed to create test app directory: %v", err)
	}
	
	cmd := exec.Command("git", "init")
	cmd.Dir = testAppPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}
	
	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testAppPath
	cmd.Run()
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testAppPath
	cmd.Run()
	
	// Cleanup on test completion
	t.Cleanup(func() {
		os.RemoveAll(testAppPath)
	})
	
	return testApp
}

func writeTestFile(t *testing.T, repo, filename, content string) {
	appsDir, _ := config.EnsureAppsDirectory()
	testAppPath := filepath.Join(appsDir, repo)
	filePath := filepath.Join(testAppPath, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

func readTestFile(t *testing.T, repo, filename string) string {
	appsDir, _ := config.EnsureAppsDirectory()
	testAppPath := filepath.Join(appsDir, repo)
	filePath := filepath.Join(testAppPath, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	return string(content)
}

func runGitCommand(t *testing.T, repo string, args ...string) string {
	appsDir, _ := config.EnsureAppsDirectory()
	testAppPath := filepath.Join(appsDir, repo)
	cmd := exec.Command("git", args...)
	cmd.Dir = testAppPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git command failed: %v, output: %s", err, string(output))
	}
	return string(output)
}

func getCommitHash(t *testing.T, repo string) string {
	output := runGitCommand(t, repo, "rev-parse", "--short", "HEAD")
	return strings.TrimSpace(output)
}

func assertFileExists(t *testing.T, repo, filename string) {
	appsDir, _ := config.EnsureAppsDirectory()
	testAppPath := filepath.Join(appsDir, repo)
	filePath := filepath.Join(testAppPath, filename)
	_, err := os.Stat(filePath)
	if err != nil {
		t.Errorf("file should exist: %s, error: %v", filename, err)
	}
}

func TestReset(t *testing.T) {
	t.Run("basic reset operations", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create initial commit
		writeTestFile(t, repo, "file1.txt", "initial content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial commit")
		
		// Get initial commit hash
		initialCommit := getCommitHash(t, repo)
		
		// Create second commit
		writeTestFile(t, repo, "file1.txt", "modified content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Second commit")
		
		// Test hard reset
		output, err := Reset(repo, initialCommit, ResetModeHard)
		if err != nil {
			t.Fatalf("reset failed: %v", err)
		}
		if !strings.Contains(output, "Successfully reset to commit") {
			t.Errorf("expected output to contain %q, got %q", "Successfully reset to commit", output)
		}
		if !strings.Contains(output, "hard mode") {
			t.Errorf("expected output to contain %q, got %q", "hard mode", output)
		}
		
		// Verify file content was reset
		content := readTestFile(t, repo, "file1.txt")
		if content != "initial content" {
			t.Errorf("expected content to be %q, got %q", "initial content", content)
		}
	})

	t.Run("soft reset keeps changes staged", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create initial commit
		writeTestFile(t, repo, "file1.txt", "initial content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial commit")
		initialCommit := getCommitHash(t, repo)
		
		// Create second commit
		writeTestFile(t, repo, "file1.txt", "modified content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Second commit")
		
		// Soft reset
		output, err := Reset(repo, initialCommit, ResetModeSoft)
		if err != nil {
			t.Fatalf("reset failed: %v", err)
		}
		if !strings.Contains(output, "soft mode") {
			t.Errorf("expected output to contain %q, got %q", "soft mode", output)
		}
		
		// Check that changes are staged
		status := runGitCommand(t, repo, "status", "--porcelain")
		if !strings.Contains(status, "M  file1.txt") {
			t.Errorf("expected status to show staged file, got %q", status)
		}
	})

	t.Run("mixed reset (default) unstages changes", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create initial commit
		writeTestFile(t, repo, "file1.txt", "initial content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial commit")
		initialCommit := getCommitHash(t, repo)
		
		// Create second commit
		writeTestFile(t, repo, "file1.txt", "modified content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Second commit")
		
		// Mixed reset (default)
		output, err := Reset(repo, initialCommit, "")
		if err != nil {
			t.Fatalf("reset failed: %v", err)
		}
		if !strings.Contains(output, "mixed mode") {
			t.Errorf("expected output to contain %q, got %q", "mixed mode", output)
		}
		
		// Check that changes are unstaged
		status := runGitCommand(t, repo, "status", "--porcelain")
		if !strings.Contains(status, " M file1.txt") {
			t.Errorf("expected status to show unstaged file, got %q", status)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing app name
		_, err := Reset("", "abc123", ResetModeHard)
		if err == nil {
			t.Error("expected error for missing app name")
		}
		if !strings.Contains(err.Error(), "app_name is required") {
			t.Errorf("expected error to contain %q, got %q", "app_name is required", err.Error())
		}
		
		// Missing commit hash
		_, err = Reset("test-app", "", ResetModeHard)
		if err == nil {
			t.Error("expected error for missing commit hash")
		}
		if !strings.Contains(err.Error(), "commit_hash is required") {
			t.Errorf("expected error to contain %q, got %q", "commit_hash is required", err.Error())
		}
		
		// Invalid mode
		_, err = Reset("test-app", "abc123", "invalid")
		if err == nil {
			t.Error("expected error for invalid mode")
		}
		if !strings.Contains(err.Error(), "invalid reset mode") {
			t.Errorf("expected error to contain %q, got %q", "invalid reset mode", err.Error())
		}
		
		// Invalid commit hash
		repo := setupTestRepo(t)
		_, err = Reset(repo, "invalid-hash", ResetModeHard)
		if err == nil {
			t.Error("expected error for invalid commit hash")
		}
	})
}