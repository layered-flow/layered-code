package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitStash(t *testing.T) {
	appsDir, _ := config.EnsureAppsDirectory()
	testApp := "test-git-stash-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath)

	os.MkdirAll(testAppPath, 0755)

	// Non-git repo
	result, err := GitStash(testApp, "push", "")
	if err != nil {
		t.Fatalf("GitStash failed: %v", err)
	}
	if result.IsRepo {
		t.Error("Expected IsRepo to be false for non-git repo")
	}

	// Setup repo with commit
	GitInit(testApp, false)
	// Configure git user for commits
	cmd := exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testAppPath
	cmd.Run()
	testFile := filepath.Join(testAppPath, "test.txt")
	os.WriteFile(testFile, []byte("original"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = testAppPath
	cmd.Run()

	// List empty stash
	result, err = GitStash(testApp, "list", "")
	if err != nil {
		t.Fatalf("GitStash list failed: %v", err)
	}
	if len(result.Stashes) != 0 {
		t.Error("Expected empty stash list")
	}

	// Modify file
	os.WriteFile(testFile, []byte("modified"), 0644)

	// Push stash
	result, err = GitStash(testApp, "push", "test stash")
	if err != nil {
		t.Fatalf("GitStash push failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}

	// Check file reverted
	content, _ := os.ReadFile(testFile)
	if string(content) != "original" {
		t.Error("Expected file to be reverted")
	}

	// Pop stash
	result, err = GitStash(testApp, "pop", "")
	if err != nil {
		t.Fatalf("GitStash pop failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}

	// Check file restored
	content, _ = os.ReadFile(testFile)
	if string(content) != "modified" {
		t.Error("Expected file to be restored")
	}
}