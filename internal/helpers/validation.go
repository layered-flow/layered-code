package helpers

import (
	"fmt"
	"strings"
)

// ValidateAppName validates that an app name is safe to use
func ValidateAppName(appName string) error {
	if appName == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	
	// Check for path traversal attempts first (before checking for period prefix)
	if strings.Contains(appName, "..") {
		return fmt.Errorf("app name cannot contain '..'")
	}
	
	// Check for hidden directories (starting with period)
	if strings.HasPrefix(appName, ".") {
		return fmt.Errorf("app name cannot start with a period (hidden directories are not allowed)")
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