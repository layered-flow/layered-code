package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitInit(t *testing.T) {
	appsDir, _ := config.EnsureAppsDirectory()
	testApp := "test-git-init-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath)

	// New directory
	result, err := GitInit(testApp, false)
	if err != nil {
		t.Fatalf("GitInit failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success")
	}

	// Check .git exists
	gitDir := filepath.Join(testAppPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		t.Error("Expected .git directory to exist")
	}

	// Already exists
	result, err = GitInit(testApp, false)
	if err != nil {
		t.Fatalf("GitInit on existing failed: %v", err)
	}
	if !result.AlreadyExists {
		t.Error("Expected AlreadyExists to be true")
	}
}