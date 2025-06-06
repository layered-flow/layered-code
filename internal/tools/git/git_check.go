package git

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
	gitAvailable     bool
	gitAvailableOnce sync.Once
	gitCheckError    error
)

// CheckGitAvailable checks if git is installed and available in PATH
func CheckGitAvailable() error {
	gitAvailableOnce.Do(func() {
		cmd := exec.Command("git", "--version")
		if err := cmd.Run(); err != nil {
			gitAvailable = false
			gitCheckError = fmt.Errorf("git is not installed or not available in PATH. Please install git to use git-related tools")
		} else {
			gitAvailable = true
			gitCheckError = nil
		}
	})
	
	return gitCheckError
}

// EnsureGitAvailable returns an error if git is not available
func EnsureGitAvailable() error {
	if err := CheckGitAvailable(); err != nil {
		return err
	}
	return nil
}