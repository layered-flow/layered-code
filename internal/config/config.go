package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/constants"
)

var (
	ErrDirectoryTraversal    = errors.New("apps directory path contains directory traversal sequences")
	ErrOutsideHomeDirectory  = errors.New("apps directory path is outside the user's home directory")
	ErrAbsolutePathNotInHome = errors.New("absolute apps directory path must be within the user's home directory")
	ErrNotADirectory         = errors.New("path is not a directory")
	ErrNotWritable           = errors.New("directory is not writable: missing write permission")
	ErrSymlinkResolution     = errors.New("failed to resolve symlinks in path")
)

// GetAppsDirectory returns the apps directory path.
// It first checks the LAYERED_APPS_DIRECTORY environment variable,
// and if not set, defaults to ~/LayeredApps
func GetAppsDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appsDir := os.Getenv(constants.AppsDirectoryEnvVar)
	if appsDir == "" {
		// Use default directory if not set
		appsDir = filepath.Join(homeDir, constants.DefaultAppsDirectory)
	} else {
		// Security validation: prevent traversal attacks
		if err := validateAppsDirectoryPath(appsDir, homeDir); err != nil {
			return "", err
		}

		// Expand ~ in the path if present
		if strings.HasPrefix(appsDir, "~/") {
			appsDir = filepath.Join(homeDir, appsDir[2:])
		} else if !filepath.IsAbs(appsDir) {
			// Convert relative paths to be relative to home directory
			appsDir = filepath.Join(homeDir, appsDir)
		}
	}

	// Clean the path
	appsDir = filepath.Clean(appsDir)

	// Final security check: ensure the resolved path is within the user's home directory
	if !isWithinDirectory(appsDir, homeDir) {
		return "", ErrOutsideHomeDirectory
	}

	return appsDir, nil
}

// EnsureAppsDirectory creates the apps directory if it doesn't exist and verifies it's writable
// Returns the apps directory path
func EnsureAppsDirectory() (string, error) {
	appsDir, err := GetAppsDirectory()
	if err != nil {
		return "", err
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(appsDir, constants.AppsDirectoryPerms); err != nil {
		return "", err
	}

	// Check if we can write to the directory
	if err := checkDirectoryWritable(appsDir); err != nil {
		return "", err
	}

	return appsDir, nil
}

// resolveSymlinks safely resolves all symlinks in a path and validates the result
func resolveSymlinks(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If the path doesn't exist yet, that's okay - we'll create it
		if os.IsNotExist(err) {
			return path, nil
		}
		return "", ErrSymlinkResolution
	}
	return resolved, nil
}

// isWithinDirectory checks if the target path is within the base directory
func isWithinDirectory(targetPath, baseDir string) bool {
	// Resolve symlinks first to prevent bypass attacks
	resolvedTarget, err := resolveSymlinks(targetPath)
	if err != nil {
		return false
	}
	resolvedBase, err := resolveSymlinks(baseDir)
	if err != nil {
		return false
	}

	// Convert both paths to absolute and clean them
	absTarget, err := filepath.Abs(resolvedTarget)
	if err != nil {
		return false
	}
	absBase, err := filepath.Abs(resolvedBase)
	if err != nil {
		return false
	}

	// Use filepath.Rel to check if target is within base
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", it's outside the base directory
	return !strings.HasPrefix(rel, "..") && rel != ".."
}

// validateAppsDirectoryPath validates the user-provided apps directory path for security
func validateAppsDirectoryPath(appsDir, homeDir string) error {
	// Prevent common traversal patterns
	if strings.Contains(appsDir, "..") {
		return ErrDirectoryTraversal
	}

	// Prevent absolute paths to system directories (allow only home-relative paths)
	if filepath.IsAbs(appsDir) {
		// Allow absolute paths only if they're within the home directory
		cleanPath := filepath.Clean(appsDir)
		if !isWithinDirectory(cleanPath, homeDir) {
			return ErrAbsolutePathNotInHome
		}
	}

	return nil
}

// checkDirectoryWritable verifies that we can write files to the specified directory
func checkDirectoryWritable(dirPath string) error {
	// Check if directory exists and get info
	info, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	// Ensure it's actually a directory
	if !info.IsDir() {
		return ErrNotADirectory
	}

	// Check write permission using file mode bits
	if info.Mode().Perm()&constants.OwnerWritePermission == 0 {
		return ErrNotWritable
	}

	return nil
}
