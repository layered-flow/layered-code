package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/layered-flow/layered-code/internal/config"
)

// Test helper functions
func setupTestRepo(t *testing.T) string {
	appsDir, err := config.EnsureAppsDirectory()
	require.NoError(t, err)
	
	testApp := fmt.Sprintf("test-git-%d", os.Getpid())
	testAppPath := filepath.Join(appsDir, testApp)
	
	// Clean up any existing directory
	os.RemoveAll(testAppPath)
	
	// Create directory and init git
	require.NoError(t, os.MkdirAll(testAppPath, 0755))
	
	cmd := exec.Command("git", "init")
	cmd.Dir = testAppPath
	require.NoError(t, cmd.Run())
	
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
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
}

func readTestFile(t *testing.T, repo, filename string) string {
	appsDir, _ := config.EnsureAppsDirectory()
	testAppPath := filepath.Join(appsDir, repo)
	filePath := filepath.Join(testAppPath, filename)
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	return string(content)
}

func runGitCommand(t *testing.T, repo string, args ...string) string {
	appsDir, _ := config.EnsureAppsDirectory()
	testAppPath := filepath.Join(appsDir, repo)
	cmd := exec.Command("git", args...)
	cmd.Dir = testAppPath
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git command failed: %s", string(output))
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
	require.NoError(t, err, "file should exist: %s", filename)
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
		output, err := Reset(repo, initialCommit, ResetModeHard, nil)
		require.NoError(t, err)
		assert.Contains(t, output, "Successfully reset to commit")
		assert.Contains(t, output, "hard mode")
		
		// Verify file content was reset
		content := readTestFile(t, repo, "file1.txt")
		assert.Equal(t, "initial content", content)
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
		output, err := Reset(repo, initialCommit, ResetModeSoft, nil)
		require.NoError(t, err)
		assert.Contains(t, output, "soft mode")
		
		// Check that changes are staged
		status := runGitCommand(t, repo, "status", "--porcelain")
		assert.Contains(t, status, "M  file1.txt")
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
		output, err := Reset(repo, initialCommit, "", nil)
		require.NoError(t, err)
		assert.Contains(t, output, "mixed mode")
		
		// Check that changes are unstaged
		status := runGitCommand(t, repo, "status", "--porcelain")
		assert.Contains(t, status, " M file1.txt")
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing app name
		_, err := Reset("", "abc123", ResetModeHard, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "app_name is required")
		
		// Missing commit hash
		_, err = Reset("test-app", "", ResetModeHard, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "commit_hash is required")
		
		// Invalid mode
		_, err = Reset("test-app", "abc123", "invalid", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid reset mode")
		
		// Invalid commit hash
		repo := setupTestRepo(t)
		_, err = Reset(repo, "invalid-hash", ResetModeHard, nil)
		assert.Error(t, err)
	})
}