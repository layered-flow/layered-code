package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitRestore(t *testing.T) {
	appsDir, _ := config.EnsureAppsDirectory()
	testApp := "test-git-restore-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath)

	os.MkdirAll(testAppPath, 0755)

	// Non-git repo
	result, err := GitRestore(testApp, []string{"file.txt"}, false)
	if err != nil {
		t.Fatalf("GitRestore failed: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for non-git repo")
	}

	// Empty files
	GitInit(testApp, false)
	// Configure git user for commits
	cmd := exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testAppPath
	cmd.Run()
	result, err = GitRestore(testApp, []string{}, false)
	if err != nil {
		t.Fatalf("GitRestore failed: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for empty files")
	}

	// Setup: create and commit a file
	testFile := filepath.Join(testAppPath, "test.txt")
	os.WriteFile(testFile, []byte("original"), 0644)
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = testAppPath
	cmd.Run()

	// Modify file
	os.WriteFile(testFile, []byte("modified"), 0644)

	// Restore working directory
	result, err = GitRestore(testApp, []string{"test.txt"}, false)
	if err != nil {
		t.Fatalf("GitRestore failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}

	// Check file was restored
	content, _ := os.ReadFile(testFile)
	if string(content) != "original" {
		t.Error("Expected file to be restored")
	}
}