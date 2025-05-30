package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitAdd(t *testing.T) {
	appsDir, _ := config.EnsureAppsDirectory()
	testApp := "test-git-add-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath)

	os.MkdirAll(testAppPath, 0755)

	// Non-git repo
	result, err := GitAdd(testApp, []string{"file.txt"}, false)
	if err != nil {
		t.Fatalf("GitAdd failed: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for non-git repo")
	}

	// Empty files list
	GitInit(testApp, false)
	// Configure git user for commits
	cmd := exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testAppPath
	cmd.Run()
	result, err = GitAdd(testApp, []string{}, false)
	if err != nil {
		t.Fatalf("GitAdd failed: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for empty files list")
	}

	// Invalid file path
	result, err = GitAdd(testApp, []string{"../outside.txt"}, false)
	if err != nil {
		t.Fatalf("GitAdd failed: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for path traversal")
	}

	// Valid add
	testFile := filepath.Join(testAppPath, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	
	result, err = GitAdd(testApp, []string{"test.txt"}, false)
	if err != nil {
		t.Fatalf("GitAdd failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}

	// Add all
	os.WriteFile(filepath.Join(testAppPath, "test2.txt"), []byte("test2"), 0644)
	result, err = GitAdd(testApp, nil, true)
	if err != nil {
		t.Fatalf("GitAdd all failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success for add all")
	}
}