package git

import (
	"fmt"
	"path/filepath"
	"strings"
	
	"github.com/layered-flow/layered-code/internal/config"
)

// ValidateAppPath validates that the app path is safe and doesn't escape the apps directory
func ValidateAppPath(appPath string) error {
	// Get the apps directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return fmt.Errorf("failed to get apps directory: %w", err)
	}
	
	// Clean both paths to resolve any ../ or ./ segments
	cleanAppPath := filepath.Clean(appPath)
	cleanAppsDir := filepath.Clean(appsDir)
	
	// Ensure the app path doesn't escape the apps directory
	// Check if the app path starts with the apps directory
	if !strings.HasPrefix(cleanAppPath, cleanAppsDir) {
		return fmt.Errorf("invalid app path: potential directory traversal detected")
	}
	
	// Additional check: ensure the relative path from appsDir to appPath doesn't contain ".."
	relPath, err := filepath.Rel(cleanAppsDir, cleanAppPath)
	if err != nil {
		return fmt.Errorf("invalid app path: %w", err)
	}
	
	// Check if the relative path tries to escape using ".."
	if strings.Contains(relPath, "..") {
		return fmt.Errorf("invalid app path: potential directory traversal detected")
	}

	return nil
}