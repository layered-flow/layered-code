package git

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckout(t *testing.T) {
	t.Run("checkout existing branch", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create initial commit on main/master
		writeTestFile(t, repo, "initial.txt", "initial content")
		runGitCommand(t, repo, "add", "initial.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial commit")
		
		// Get the default branch name (could be main or master)
		defaultBranch := strings.TrimSpace(runGitCommand(t, repo, "branch", "--show-current"))
		
		// Create a branch
		runGitCommand(t, repo, "checkout", "-b", "feature-branch")
		writeTestFile(t, repo, "feature.txt", "feature content")
		runGitCommand(t, repo, "add", "feature.txt")
		runGitCommand(t, repo, "commit", "-m", "Add feature")
		
		// Switch back to default branch
		runGitCommand(t, repo, "checkout", defaultBranch)
		
		// Use our checkout function
		output, err := Checkout(repo, "feature-branch", false, nil, nil)
		require.NoError(t, err)
		assert.Contains(t, output, "Successfully checked out branch/commit 'feature-branch'")
		assert.Contains(t, output, "Current branch: feature-branch")
		
		// Verify we're on feature branch
		branch := runGitCommand(t, repo, "branch", "--show-current")
		assert.Contains(t, branch, "feature-branch")
		
		// Verify feature file exists
		assertFileExists(t, repo, "feature.txt")
	})

	t.Run("checkout new branch", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create new branch
		output, err := Checkout(repo, "new-feature", true, nil, nil)
		require.NoError(t, err)
		assert.Contains(t, output, "Successfully checked out new branch 'new-feature'")
		assert.Contains(t, output, "Current branch: new-feature")
		
		// Verify we're on the new branch
		branch := runGitCommand(t, repo, "branch", "--show-current")
		assert.Contains(t, branch, "new-feature")
	})

	t.Run("checkout specific commit", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create some commits
		writeTestFile(t, repo, "file1.txt", "v1")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "First commit")
		firstCommit := getCommitHash(t, repo)
		
		writeTestFile(t, repo, "file1.txt", "v2")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Second commit")
		
		// Checkout first commit
		output, err := Checkout(repo, firstCommit, false, nil, nil)
		require.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Successfully checked out branch/commit '%s'", firstCommit))
		assert.Contains(t, output, "HEAD is now at")
		assert.Contains(t, output, "(detached)")
		
		// Verify file content
		content := readTestFile(t, repo, "file1.txt")
		assert.Equal(t, "v1", content)
	})

	t.Run("checkout specific files from HEAD", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create and commit a file
		writeTestFile(t, repo, "file1.txt", "original")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial")
		
		// Modify the file (don't commit)
		writeTestFile(t, repo, "file1.txt", "modified")
		
		// Checkout file from HEAD
		output, err := Checkout(repo, "", false, []string{"file1.txt"}, nil)
		require.NoError(t, err)
		assert.Contains(t, output, "Successfully checked out files from HEAD")
		
		// Verify file was restored
		content := readTestFile(t, repo, "file1.txt")
		assert.Equal(t, "original", content)
	})

	t.Run("checkout files from specific commit", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// First commit
		writeTestFile(t, repo, "file1.txt", "v1")
		writeTestFile(t, repo, "file2.txt", "v1")
		runGitCommand(t, repo, "add", ".")
		runGitCommand(t, repo, "commit", "-m", "First commit")
		firstCommit := getCommitHash(t, repo)
		
		// Second commit
		writeTestFile(t, repo, "file1.txt", "v2")
		writeTestFile(t, repo, "file2.txt", "v2")
		runGitCommand(t, repo, "add", ".")
		runGitCommand(t, repo, "commit", "-m", "Second commit")
		
		// Checkout file1 from first commit
		output, err := Checkout(repo, firstCommit, false, []string{"file1.txt"}, nil)
		require.NoError(t, err)
		assert.Contains(t, output, fmt.Sprintf("Successfully checked out files from %s", firstCommit))
		
		// Verify file1 is v1, file2 is still v2
		content1 := readTestFile(t, repo, "file1.txt")
		content2 := readTestFile(t, repo, "file2.txt")
		assert.Equal(t, "v1", content1)
		assert.Equal(t, "v2", content2)
	})

	t.Run("checkout multiple files", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create and commit files
		writeTestFile(t, repo, "file1.txt", "original1")
		writeTestFile(t, repo, "file2.txt", "original2")
		writeTestFile(t, repo, "file3.txt", "original3")
		runGitCommand(t, repo, "add", ".")
		runGitCommand(t, repo, "commit", "-m", "Initial")
		
		// Modify all files
		writeTestFile(t, repo, "file1.txt", "modified1")
		writeTestFile(t, repo, "file2.txt", "modified2")
		writeTestFile(t, repo, "file3.txt", "modified3")
		
		// Checkout only file1 and file2
		output, err := Checkout(repo, "", false, []string{"file1.txt", "file2.txt"}, nil)
		require.NoError(t, err)
		assert.Contains(t, output, "Successfully checked out files from HEAD")
		
		// Verify file1 and file2 restored, file3 still modified
		assert.Equal(t, "original1", readTestFile(t, repo, "file1.txt"))
		assert.Equal(t, "original2", readTestFile(t, repo, "file2.txt"))
		assert.Equal(t, "modified3", readTestFile(t, repo, "file3.txt"))
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing app name
		_, err := Checkout("", "main", false, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "app_name is required")
		
		// No target or files
		_, err = Checkout("test-app", "", false, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "either target branch/commit or files must be specified")
		
		// Non-existent branch
		repo := setupTestRepo(t)
		_, err = Checkout(repo, "non-existent-branch", false, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git checkout failed")
		
		// Non-existent file
		_, err = Checkout(repo, "", false, []string{"non-existent.txt"}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "git checkout failed")
	})
}