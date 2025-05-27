package tools

import (
	"os/exec"
	"testing"
)

func TestCheckGitAvailable(t *testing.T) {
	// This test assumes git is installed on the test system
	err := CheckGitAvailable()
	
	// Check if git command exists
	_, gitErr := exec.LookPath("git")
	
	if gitErr != nil {
		// Git is not installed
		if err == nil {
			t.Error("Expected error when git is not installed, but got nil")
		}
		if err.Error() != "git is not installed or not available in PATH. Please install git to use git-related tools" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	} else {
		// Git is installed
		if err != nil {
			t.Errorf("Expected no error when git is installed, but got: %v", err)
		}
	}
}

func TestGitToolsWithoutGit(t *testing.T) {
	// Skip this test if git is actually installed
	if _, err := exec.LookPath("git"); err == nil {
		t.Skip("Skipping test because git is installed")
	}
	
	// Test that git tools return appropriate errors when git is not available
	t.Run("GitStatus", func(t *testing.T) {
		_, err := GitStatus("testapp")
		if err == nil {
			t.Error("Expected error when git is not available")
		}
		if err.Error() != "git is not installed or not available in PATH. Please install git to use git-related tools" {
			t.Errorf("Expected git not available error, got: %v", err)
		}
	})
	
	t.Run("GitDiff", func(t *testing.T) {
		_, err := GitDiff("testapp", false, "")
		if err == nil {
			t.Error("Expected error when git is not available")
		}
	})
	
	t.Run("GitInit", func(t *testing.T) {
		_, err := GitInit("testapp", false)
		if err == nil {
			t.Error("Expected error when git is not available")
		}
	})
}