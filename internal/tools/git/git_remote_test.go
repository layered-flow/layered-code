package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestGitRemote(t *testing.T) {
	// Setup test environment
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create a test app directory
	testApp := "test-git-remote-app"
	testAppPath := filepath.Join(appsDir, testApp)
	defer os.RemoveAll(testAppPath) // Cleanup

	// Create app directory
	if err := os.MkdirAll(testAppPath, 0755); err != nil {
		t.Fatalf("Failed to create test app directory: %v", err)
	}

	// Initialize a git repository
	_, err = GitInit(testApp, false)
	if err != nil {
		t.Fatalf("Failed to initialize git repository: %v", err)
	}

	t.Run("list remotes in empty repo", func(t *testing.T) {
		result, err := GitRemote(testApp, "", "", "", "", "", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if len(result.Remotes) != 0 {
			t.Errorf("Expected no remotes, got %d", len(result.Remotes))
		}
	})

	t.Run("add remote", func(t *testing.T) {
		result, err := GitRemote(testApp, "origin", "https://github.com/user/repo.git", "", "", "", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if !result.AddSuccess {
			t.Error("Expected AddSuccess to be true")
		}
		if _, ok := result.Remotes["origin"]; !ok {
			t.Error("Expected 'origin' remote to be present")
		}
		if result.Remotes["origin"] != "https://github.com/user/repo.git" {
			t.Errorf("Expected remote URL to be 'https://github.com/user/repo.git', got '%s'", result.Remotes["origin"])
		}
	})

	t.Run("add duplicate remote fails", func(t *testing.T) {
		result, err := GitRemote(testApp, "origin", "https://github.com/user/other.git", "", "", "", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if result.AddSuccess {
			t.Error("Expected AddSuccess to be false for duplicate remote")
		}
		if result.Message == "" {
			t.Error("Expected error message for duplicate remote")
		}
	})

	t.Run("add second remote", func(t *testing.T) {
		result, err := GitRemote(testApp, "upstream", "https://github.com/upstream/repo.git", "", "", "", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if !result.AddSuccess {
			t.Error("Expected AddSuccess to be true")
		}
		if len(result.Remotes) != 2 {
			t.Errorf("Expected 2 remotes, got %d", len(result.Remotes))
		}
		if _, ok := result.Remotes["upstream"]; !ok {
			t.Error("Expected 'upstream' remote to be present")
		}
	})

	t.Run("set remote URL", func(t *testing.T) {
		result, err := GitRemote(testApp, "", "", "", "", "", "https://github.com/user/new-repo.git", "origin")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if result.Remotes["origin"] != "https://github.com/user/new-repo.git" {
			t.Errorf("Expected remote URL to be updated to 'https://github.com/user/new-repo.git', got '%s'", result.Remotes["origin"])
		}
	})

	t.Run("rename remote", func(t *testing.T) {
		result, err := GitRemote(testApp, "", "", "", "upstream", "backup", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if !result.RenameSuccess {
			t.Error("Expected RenameSuccess to be true")
		}
		if _, ok := result.Remotes["backup"]; !ok {
			t.Error("Expected 'backup' remote to be present")
		}
		if _, ok := result.Remotes["upstream"]; ok {
			t.Error("Expected 'upstream' remote to be absent after rename")
		}
	})

	t.Run("remove remote", func(t *testing.T) {
		result, err := GitRemote(testApp, "", "", "backup", "", "", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if !result.RemoveSuccess {
			t.Error("Expected RemoveSuccess to be true")
		}
		if _, ok := result.Remotes["backup"]; ok {
			t.Error("Expected 'backup' remote to be absent after removal")
		}
		if len(result.Remotes) != 1 {
			t.Errorf("Expected 1 remote remaining, got %d", len(result.Remotes))
		}
	})

	t.Run("remove non-existent remote fails", func(t *testing.T) {
		result, err := GitRemote(testApp, "", "", "nonexistent", "", "", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if !result.IsRepo {
			t.Error("Expected IsRepo to be true")
		}
		if result.RemoveSuccess {
			t.Error("Expected RemoveSuccess to be false for non-existent remote")
		}
		if result.Message == "" {
			t.Error("Expected error message for non-existent remote")
		}
	})

	t.Run("non-git directory", func(t *testing.T) {
		nonGitApp := "non-git-app"
		appPath := filepath.Join(appsDir, nonGitApp)
		err := os.MkdirAll(appPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create non-git app directory: %v", err)
		}
		defer os.RemoveAll(appPath)

		result, err := GitRemote(nonGitApp, "", "", "", "", "", "", "")
		if err != nil {
			t.Fatalf("GitRemote failed: %v", err)
		}
		if result.IsRepo {
			t.Error("Expected IsRepo to be false for non-git directory")
		}
		if result.Message == "" {
			t.Error("Expected message for non-git repository")
		}
	})

	t.Run("invalid app name", func(t *testing.T) {
		_, err := GitRemote("../invalid", "", "", "", "", "", "", "")
		if err == nil {
			t.Error("Expected error for invalid app name")
		}
	})
}