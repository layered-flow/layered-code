package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateAppName validates the app name to ensure it's safe
func ValidateAppName(appName string) error {
	if appName == "" {
		return fmt.Errorf("app_name is required")
	}

	// Check for path traversal attempts
	if strings.Contains(appName, "..") || strings.Contains(appName, "/") || strings.Contains(appName, "\\") {
		return fmt.Errorf("invalid app_name: must not contain path separators or '..'")
	}

	return nil
}

// ValidateAppPath validates that the app path is safe and doesn't escape the apps directory
func ValidateAppPath(appPath string) error {
	// Clean the path to resolve any ../ or ./ segments
	cleanPath := filepath.Clean(appPath)
	
	// Ensure the path doesn't escape the apps directory
	if !strings.HasPrefix(cleanPath, filepath.Clean(appPath)) {
		return fmt.Errorf("invalid app path: potential directory traversal detected")
	}

	return nil
}