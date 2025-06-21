package git

import (
	"strings"
	"testing"
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
		if err != nil {
			t.Fatalf("revert failed: %v", err)
		}
		if !strings.Contains(output, "Successfully created revert commit") {
			t.Errorf("expected output to contain %q, got %q", "Successfully created revert commit", output)
		}
		if !strings.Contains(output, "New commit:") {
			t.Errorf("expected output to contain %q, got %q", "New commit:", output)
		}
		
		// Verify file content was reverted
		content := readTestFile(t, repo, "file1.txt")
		if content != "initial content" {
			t.Errorf("expected content to be %q, got %q", "initial content", content)
		}
		
		// Verify a new commit was created
		log := runGitCommand(t, repo, "log", "--oneline", "-3")
		if !strings.Contains(log, "Revert") {
			t.Errorf("expected log to contain %q, got %q", "Revert", log)
		}
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
		if err != nil {
			t.Fatalf("revert failed: %v", err)
		}
		if !strings.Contains(output, "changes staged but not committed") {
			t.Errorf("expected output to contain %q, got %q", "changes staged but not committed", output)
		}
		
		// Verify changes are staged
		status := runGitCommand(t, repo, "status", "--porcelain")
		if !strings.Contains(status, "M  file1.txt") {
			t.Errorf("expected status to show staged file, got %q", status)
		}
		
		// Verify no new commit was created
		currentHash := getCommitHash(t, repo)
		if currentHash != commitToRevert {
			t.Errorf("expected current hash to be %q, got %q", commitToRevert, currentHash)
		}
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
		if err != nil {
			t.Fatalf("revert failed: %v", err)
		}
		if !strings.Contains(output, "Successfully created revert commit") {
			t.Errorf("expected output to contain %q, got %q", "Successfully created revert commit", output)
		}
		
		// Verify both files were reverted
		content1 := readTestFile(t, repo, "file1.txt")
		content2 := readTestFile(t, repo, "file2.txt")
		if content1 != "content1" {
			t.Errorf("expected file1 content to be %q, got %q", "content1", content1)
		}
		if content2 != "content2" {
			t.Errorf("expected file2 content to be %q, got %q", "content2", content2)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		// Missing app name
		_, err := Revert("", "abc123", false)
		if err == nil {
			t.Error("expected error for missing app name")
		}
		if !strings.Contains(err.Error(), "app_name is required") {
			t.Errorf("expected error to contain %q, got %q", "app_name is required", err.Error())
		}
		
		// Missing commit hash
		_, err = Revert("test-app", "", false)
		if err == nil {
			t.Error("expected error for missing commit hash")
		}
		if !strings.Contains(err.Error(), "commit_hash is required") {
			t.Errorf("expected error to contain %q, got %q", "commit_hash is required", err.Error())
		}
		
		// Invalid commit hash
		repo := setupTestRepo(t)
		writeTestFile(t, repo, "file1.txt", "content")
		runGitCommand(t, repo, "add", "file1.txt")
		runGitCommand(t, repo, "commit", "-m", "Initial")
		
		_, err = Revert(repo, "invalid-hash", false)
		if err == nil {
			t.Error("expected error for invalid commit hash")
		}
		if !strings.Contains(err.Error(), "git revert failed") {
			t.Errorf("expected error to contain %q, got %q", "git revert failed", err.Error())
		}
	})
}