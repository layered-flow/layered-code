package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevert(t *testing.T) {
	t.Run("revert with commit", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create initial commit
		writeTestFile(t, repo, "file1.txt", "initial content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial commit")
		
		// Create second commit that we'll revert
		writeTestFile(t, repo, "file1.txt", "modified content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Change to revert")
		commitToRevert := getCommitHash(t, repo)
		
		// Revert the commit
		output, err := Revert(repo, commitToRevert, false)
		require.NoError(t, err)
		assert.Contains(t, output, "Successfully created revert commit")
		assert.Contains(t, output, "New commit:")
		
		// Verify file content was reverted
		content := readTestFile(t, repo, "file1.txt")
		assert.Equal(t, "initial content", content)
		
		// Verify a new commit was created
		log := runGitCommand(t, repo, "log", "--oneline", "-3")
		assert.Contains(t, log, "Revert")
	})

	t.Run("revert without commit", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create initial commit
		writeTestFile(t, repo, "file1.txt", "initial content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial commit")
		
		// Create second commit
		writeTestFile(t, repo, "file1.txt", "modified content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Change to revert")
		commitToRevert := getCommitHash(t, repo)
		
		// Revert without committing
		output, err := Revert(repo, commitToRevert, true)
		require.NoError(t, err)
		assert.Contains(t, output, "changes staged but not committed")
		
		// Verify changes are staged
		status := runGitCommand(t, repo, "status", "--porcelain")
		assert.Contains(t, status, "M  file1.txt")
		
		// Verify no new commit was created
		currentHash := getCommitHash(t, repo)
		assert.Equal(t, commitToRevert, currentHash)
	})

	t.Run("revert multiple file changes", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Initial state
		writeTestFile(t, repo, "file1.txt", "content1")
		writeTestFile(t, repo, "file2.txt", "content2")
		runGitCommand(t, repo, "add", ".")
		runGitCommand(t, repo, "commit", "-m", "Initial commit")
		
		// Make changes to multiple files
		writeTestFile(t, repo, "file1.txt", "changed1")
		writeTestFile(t, repo, "file2.txt", "changed2")
		runGitCommand(t, repo, "add", ".")
		runGitCommand(t, repo, "commit", "-m", "Multi-file changes")
		commitToRevert := getCommitHash(t, repo)
		
		// Revert
		output, err := Revert(repo, commitToRevert, false)
		require.NoError(t, err)
		assert.Contains(t, output, "Successfully created revert commit")
		
		// Verify both files were reverted
		content1 := readTestFile(t, repo, "file1.txt")
		content2 := readTestFile(t, repo, "file2.txt")
		assert.Equal(t, "content1", content1)
		assert.Equal(t, "content2", content2)
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing app name
		_, err := Revert("", "abc123", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "app_name is required")
		
		// Missing commit hash
		_, err = Revert("test-app", "", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "commit_hash is required")
		
		// Invalid commit hash
		repo := setupTestRepo(t)
		writeTestFile(t, repo, "file1.txt", "content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial")
		
		_, err = Revert(repo, "invalid-hash", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git revert failed")
	})
}