package tools

import (
	"os/exec"
)

// initGitRepo initializes a git repository for testing
func initGitRepo(path string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = path
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = path
	cmd.Run()
}