package pnpm

import (
	"fmt"
	"os/exec"
	"strings"
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

// ValidateAppName validates that an app name is safe to use
func ValidateAppName(appName string) error {
	if appName == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	
	// Check for path traversal attempts
	if strings.Contains(appName, "..") {
		return fmt.Errorf("app name cannot contain '..'")
	}
	
	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(appName, char) {
			return fmt.Errorf("app name cannot contain '%s'", char)
		}
	}
	
	return nil
}