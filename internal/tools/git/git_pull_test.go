package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitPull(t *testing.T) {
	appsDir, _ := config.EnsureAppsDirectory()
	testApp := "test-git-pull-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath)

	// Test 1: Non-git repo
	os.MkdirAll(testAppPath, 0755)
	result, err := GitPull(testApp, "origin", "main", false)
	if err != nil {
		t.Fatalf("GitPull failed: %v", err)
	}
	if result.Success {
		t.Error("Expected failure for non-git repo")
	}
	if result.IsRepo {
		t.Error("Expected IsRepo to be false")
	}

	// Clean up and set up for remaining tests
	os.RemoveAll(testAppPath)

	// Set up test repository structure
	remoteDir := filepath.Join(appsDir, "test-pull-remote")
	os.MkdirAll(remoteDir, 0755)
	defer os.RemoveAll(remoteDir)
	
	// Initialize bare remote repo with initial content
	tempDir := filepath.Join(appsDir, "test-pull-temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)
	
	// Create temp repo to push initial content
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	cmd.Run()
	
	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	cmd.Run()
	
	// Create initial file
	initialFile := filepath.Join(tempDir, "README.md")
	os.WriteFile(initialFile, []byte("Initial content"), 0644)
	
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = tempDir
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tempDir
	cmd.Run()
	
	// Initialize bare repo and push
	cmd = exec.Command("git", "init", "--bare")
	cmd.Dir = remoteDir
	cmd.Run()
	
	cmd = exec.Command("git", "remote", "add", "origin", remoteDir)
	cmd.Dir = tempDir
	cmd.Run()
	
	cmd = exec.Command("git", "push", "-u", "origin", "master")
	cmd.Dir = tempDir
	cmd.Run()

	// Clone to test app directory
	cmd = exec.Command("git", "clone", remoteDir, testAppPath)
	cmd.Dir = appsDir
	cmd.Run()

	// Configure git user for test app
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testAppPath
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testAppPath
	cmd.Run()
	// Set pull strategy to merge
	cmd = exec.Command("git", "config", "pull.rebase", "false")
	cmd.Dir = testAppPath
	cmd.Run()

	// Test 2: Pull when already up to date
	result, err = GitPull(testApp, "origin", "master", false)
	if err != nil {
		t.Fatalf("GitPull failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success for pull. Error output: %s", result.ErrorOutput)
	}
	if result.Updated {
		t.Error("Expected Updated to be false when already up to date")
	}

	// Create another working directory to simulate remote changes
	workDir := filepath.Join(appsDir, "test-pull-work")
	cmd = exec.Command("git", "clone", remoteDir, workDir)
	cmd.Dir = appsDir
	cmd.Run()
	defer os.RemoveAll(workDir)

	// Configure git user in work dir
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = workDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = workDir
	cmd.Run()

	// Make changes in the work dir and push
	remoteFile := filepath.Join(workDir, "remote-change.txt")
	os.WriteFile(remoteFile, []byte("remote content"), 0644)
	
	cmd = exec.Command("git", "add", "remote-change.txt")
	cmd.Dir = workDir
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Remote change")
	cmd.Dir = workDir
	cmd.Run()
	
	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = workDir
	cmd.Run()

	// Test 3: Pull with new changes
	result, err = GitPull(testApp, "origin", "master", false)
	if err != nil {
		t.Fatalf("GitPull failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success for pull with changes. Error output: %s", result.ErrorOutput)
	}
	if !result.Updated {
		t.Error("Expected Updated to be true when pulling new changes")
	}

	// Verify the file was pulled
	pulledFile := filepath.Join(testAppPath, "remote-change.txt")
	if _, err := os.Stat(pulledFile); os.IsNotExist(err) {
		t.Error("Expected remote-change.txt to exist after pull")
	}

	// Test 4: Pull with empty remote (should default to origin)
	result, err = GitPull(testApp, "", "master", false)
	if err != nil {
		t.Fatalf("GitPull with empty remote failed: %v", err)
	}
	if !result.Success {
		t.Error("Expected success for pull with empty remote")
	}

	// Test 5: Pull with rebase
	// Make local changes
	localFile := filepath.Join(testAppPath, "local-change.txt")
	os.WriteFile(localFile, []byte("local content"), 0644)
	
	cmd = exec.Command("git", "add", "local-change.txt")
	cmd.Dir = testAppPath
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Local change")
	cmd.Dir = testAppPath
	cmd.Run()

	// Make another remote change
	remoteFile2 := filepath.Join(workDir, "another-remote.txt")
	os.WriteFile(remoteFile2, []byte("another remote content"), 0644)
	
	cmd = exec.Command("git", "add", "another-remote.txt")
	cmd.Dir = workDir
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Another remote change")
	cmd.Dir = workDir
	cmd.Run()
	
	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = workDir
	cmd.Run()

	// Pull with rebase
	result, err = GitPull(testApp, "origin", "master", true)
	if err != nil {
		t.Fatalf("GitPull with rebase failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success for pull with rebase. Error output: %s", result.ErrorOutput)
	}

	// Test 6: Pull from specific branch
	// Create and push a new branch in work dir
	cmd = exec.Command("git", "checkout", "-b", "feature-branch")
	cmd.Dir = workDir
	cmd.Run()

	branchFile := filepath.Join(workDir, "branch-file.txt")
	os.WriteFile(branchFile, []byte("branch content"), 0644)
	
	cmd = exec.Command("git", "add", "branch-file.txt")
	cmd.Dir = workDir
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Branch commit")
	cmd.Dir = workDir
	cmd.Run()
	
	cmd = exec.Command("git", "push", "origin", "feature-branch")
	cmd.Dir = workDir
	cmd.Run()

	// Fetch and create tracking branch 
	cmd = exec.Command("git", "fetch", "origin")
	cmd.Dir = testAppPath
	cmd.Run()

	cmd = exec.Command("git", "checkout", "-b", "feature-branch", "origin/feature-branch")
	cmd.Dir = testAppPath
	cmd.Run()

	result, err = GitPull(testApp, "origin", "feature-branch", false)
	if err != nil {
		t.Fatalf("GitPull from specific branch failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success for pull from branch. Error output: %s", result.ErrorOutput)
	}
}