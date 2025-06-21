package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitPush(t *testing.T) {
	appsDir, _ := config.EnsureAppsDirectory()
	testApp := "test-git-push-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath)

	os.MkdirAll(testAppPath, 0755)

	// Test 1: Non-git repo
	result, err := GitPush(testApp, "origin", "main", false, false)
	if err != nil {
		t.Fatalf("GitPush failed: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for non-git repo")
	}
	if result.IsRepo {
		t.Error("Expected IsRepo to be false")
	}

	// Initialize git repo
	GitInit(testApp, false)
	
	// Configure git user for commits
	cmd := exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testAppPath
	cmd.Run()

	// Test 2: Push without remote
	result, err = GitPush(testApp, "origin", "main", false, false)
	if err == nil {
		t.Error("Expected error for push without remote")
	}
	if result.Success {
		t.Error("Expected failure for push without remote")
	}

	// Test 3: Push with empty remote (should default to origin)
	result, err = GitPush(testApp, "", "main", false, false)
	if err == nil {
		t.Error("Expected error for push without configured remote")
	}

	// Add a test remote (using a local directory as remote)
	remoteDir := filepath.Join(appsDir, "test-remote.git")
	os.MkdirAll(remoteDir, 0755)
	defer os.RemoveAll(remoteDir)
	
	// Initialize bare remote repo
	cmd = exec.Command("git", "init", "--bare")
	cmd.Dir = remoteDir
	cmd.Run()

	// Add remote to test repo
	cmd = exec.Command("git", "remote", "add", "origin", remoteDir)
	cmd.Dir = testAppPath
	cmd.Run()

	// Create and commit a file
	testFile := filepath.Join(testAppPath, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)
	
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = testAppPath
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = testAppPath
	cmd.Run()

	// Test 4: Valid push
	// Create main branch first
	cmd = exec.Command("git", "checkout", "-b", "main")
	cmd.Dir = testAppPath
	cmd.Run()
	
	result, err = GitPush(testApp, "origin", "main", false, false)
	if err != nil {
		t.Fatalf("GitPush failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success for valid push. Error output: %s", result.ErrorOutput)
	}

	// Test 5: Push with set upstream
	cmd = exec.Command("git", "checkout", "-b", "feature-branch")
	cmd.Dir = testAppPath
	cmd.Run()

	// Make another commit
	os.WriteFile(testFile, []byte("updated content"), 0644)
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "Update file")
	cmd.Dir = testAppPath
	cmd.Run()

	result, err = GitPush(testApp, "origin", "feature-branch", true, false)
	if err != nil {
		t.Fatalf("GitPush with set-upstream failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success for push with set-upstream. Error output: %s", result.ErrorOutput)
	}

	// Test 6: Force push
	// Create a conflicting commit
	os.WriteFile(testFile, []byte("force push content"), 0644)
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "commit", "--amend", "-m", "Amended commit")
	cmd.Dir = testAppPath
	cmd.Run()

	result, err = GitPush(testApp, "origin", "feature-branch", false, true)
	if err != nil {
		t.Fatalf("GitPush with force failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success for force push. Error output: %s", result.ErrorOutput)
	}
}