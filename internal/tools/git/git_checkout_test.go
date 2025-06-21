package git

import (
	"fmt"
	"strings"
	"testing"
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
		output, err := Checkout(repo, "feature-branch", false, nil)
		if err != nil {
			t.Fatalf("checkout failed: %v", err)
		}
		if !strings.Contains(output, "Successfully checked out branch/commit 'feature-branch'") {
			t.Errorf("expected output to contain %q, got %q", "Successfully checked out branch/commit 'feature-branch'", output)
		}
		if !strings.Contains(output, "Current branch: feature-branch") {
			t.Errorf("expected output to contain %q, got %q", "Current branch: feature-branch", output)
		}
		
		// Verify we're on feature branch
		branch := runGitCommand(t, repo, "branch", "--show-current")
		if !strings.Contains(branch, "feature-branch") {
			t.Errorf("expected branch to contain %q, got %q", "feature-branch", branch)
		}
		
		// Verify feature file exists
		assertFileExists(t, repo, "feature.txt")
	})

	t.Run("checkout new branch", func(t *testing.T) {
		repo := setupTestRepo(t)
		
		// Create new branch
		output, err := Checkout(repo, "new-feature", true, nil)
		if err != nil {
			t.Fatalf("checkout failed: %v", err)
		}
		if !strings.Contains(output, "Successfully checked out new branch 'new-feature'") {
			t.Errorf("expected output to contain %q, got %q", "Successfully checked out new branch 'new-feature'", output)
		}
		if !strings.Contains(output, "Current branch: new-feature") {
			t.Errorf("expected output to contain %q, got %q", "Current branch: new-feature", output)
		}
		
		// Verify we're on the new branch
		branch := runGitCommand(t, repo, "branch", "--show-current")
		if !strings.Contains(branch, "new-feature") {
			t.Errorf("expected branch to contain %q, got %q", "new-feature", branch)
		}
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
		output, err := Checkout(repo, firstCommit, false, nil)
		if err != nil {
			t.Fatalf("checkout failed: %v", err)
		}
		expectedMsg := fmt.Sprintf("Successfully checked out branch/commit '%s'", firstCommit)
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("expected output to contain %q, got %q", expectedMsg, output)
		}
		if !strings.Contains(output, "HEAD is now at") {
			t.Errorf("expected output to contain %q, got %q", "HEAD is now at", output)
		}
		if !strings.Contains(output, "(detached)") {
			t.Errorf("expected output to contain %q, got %q", "(detached)", output)
		}
		
		// Verify file content
		content := readTestFile(t, repo, "file1.txt")
		if content != "v1" {
			t.Errorf("expected content to be %q, got %q", "v1", content)
		}
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
		output, err := Checkout(repo, "", false, []string{"file1.txt"})
		if err != nil {
			t.Fatalf("checkout failed: %v", err)
		}
		if !strings.Contains(output, "Successfully checked out files from HEAD") {
			t.Errorf("expected output to contain %q, got %q", "Successfully checked out files from HEAD", output)
		}
		
		// Verify file was restored
		content := readTestFile(t, repo, "file1.txt")
		if content != "original" {
			t.Errorf("expected content to be %q, got %q", "original", content)
		}
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
		output, err := Checkout(repo, firstCommit, false, []string{"file1.txt"})
		if err != nil {
			t.Fatalf("checkout failed: %v", err)
		}
		expectedMsg := fmt.Sprintf("Successfully checked out files from %s", firstCommit)
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("expected output to contain %q, got %q", expectedMsg, output)
		}
		
		// Verify file1 is v1, file2 is still v2
		content1 := readTestFile(t, repo, "file1.txt")
		content2 := readTestFile(t, repo, "file2.txt")
		if content1 != "v1" {
			t.Errorf("expected file1 content to be %q, got %q", "v1", content1)
		}
		if content2 != "v2" {
			t.Errorf("expected file2 content to be %q, got %q", "v2", content2)
		}
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
		output, err := Checkout(repo, "", false, []string{"file1.txt", "file2.txt"})
		if err != nil {
			t.Fatalf("checkout failed: %v", err)
		}
		if !strings.Contains(output, "Successfully checked out files from HEAD") {
			t.Errorf("expected output to contain %q, got %q", "Successfully checked out files from HEAD", output)
		}
		
		// Verify file1 and file2 restored, file3 still modified
		if content := readTestFile(t, repo, "file1.txt"); content != "original1" {
			t.Errorf("expected file1 content to be %q, got %q", "original1", content)
		}
		if content := readTestFile(t, repo, "file2.txt"); content != "original2" {
			t.Errorf("expected file2 content to be %q, got %q", "original2", content)
		}
		if content := readTestFile(t, repo, "file3.txt"); content != "modified3" {
			t.Errorf("expected file3 content to be %q, got %q", "modified3", content)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing app name
		_, err := Checkout("", "main", false, nil)
		if err == nil {
			t.Error("expected error for missing app name")
		}
		if !strings.Contains(err.Error(), "app_name is required") {
			t.Errorf("expected error to contain %q, got %q", "app_name is required", err.Error())
		}
		
		// No target or files
		_, err = Checkout("test-app", "", false, nil)
		if err == nil {
			t.Error("expected error for no target or files")
		}
		if !strings.Contains(err.Error(), "either target branch/commit or files must be specified") {
			t.Errorf("expected error to contain %q, got %q", "either target branch/commit or files must be specified", err.Error())
		}
		
		// Non-existent branch
		repo := setupTestRepo(t)
		_, err = Checkout(repo, "non-existent-branch", false, nil)
		if err == nil {
			t.Error("expected error for non-existent branch")
		}
		if !strings.Contains(err.Error(), "git checkout failed") {
			t.Errorf("expected error to contain %q, got %q", "git checkout failed", err.Error())
		}
		
		// Non-existent file
		_, err = Checkout(repo, "", false, []string{"non-existent.txt"})
		if err == nil {
			t.Error("expected error for non-existent file")
		}
		if !strings.Contains(err.Error(), "git checkout failed") {
			t.Errorf("expected error to contain %q, got %q", "git checkout failed", err.Error())
		}
	})
}