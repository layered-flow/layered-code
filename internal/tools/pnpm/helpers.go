package pnpm

import (
	"fmt"
	"os/exec"
)

// DetectPackageManager detects which package manager is available (pnpm or npm)
// Returns the package manager name or an error if neither is available
func DetectPackageManager() (string, error) {
	// Check for pnpm first (preferred)
	if _, err := exec.LookPath("pnpm"); err == nil {
		return "pnpm", nil
	}
	
	// Fall back to npm
	if _, err := exec.LookPath("npm"); err == nil {
		return "npm", nil
	}
	
	return "", fmt.Errorf("neither pnpm nor npm is available. Please install Node.js and npm or pnpm")
}