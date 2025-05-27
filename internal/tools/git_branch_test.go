package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitBranch(t *testing.T) {
	appsDir, _ := config.EnsureAppsDirectory()
	testApp := "test-git-branch-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath)

	os.MkdirAll(testAppPath, 0755)

	// Non-git repo
	result, err := GitBranch(testApp, "", "", "", false)
	if err != nil {
		t.Fatalf("GitBranch failed: %v", err)
	}
	if result.IsRepo {
		t.Error("Expected IsRepo to be false for non-git repo")
	}

	// Initialize repo with a commit
	initGitRepo(testAppPath)
	testFile := filepath.Join(testAppPath, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = testAppPath
	cmd.Run()

	// List branches
	result, err = GitBranch(testApp, "", "", "", false)
	if err != nil {
		t.Fatalf("GitBranch list failed: %v", err)
	}
	if len(result.Branches) == 0 {
		t.Error("Expected at least one branch")
	}

	// Create branch
	result, err = GitBranch(testApp, "feature", "", "", false)
	if err != nil {
		t.Fatalf("GitBranch create failed: %v", err)
	}
	if !result.CreateSuccess {
		t.Error("Expected create success")
	}

	// Switch branch
	result, err = GitBranch(testApp, "", "feature", "", false)
	if err != nil {
		t.Fatalf("GitBranch switch failed: %v", err)
	}
	if !result.SwitchSuccess {
		t.Error("Expected switch success")
	}

	// Delete branch (switch back first)
	GitBranch(testApp, "", "master", "", false)
	result, err = GitBranch(testApp, "", "", "feature", false)
	if err != nil {
		t.Fatalf("GitBranch delete failed: %v", err)
	}
	if !result.DeleteSuccess {
		t.Error("Expected delete success")
	}
}