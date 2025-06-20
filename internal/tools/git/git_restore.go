package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type GitRestoreResult struct {
	Success       bool     `json:"success"`
	FilesRestored []string `json:"files_restored"`
	IsRepo        bool     `json:"is_repo"`
	Message       string   `json:"message,omitempty"`
}

// GitRestore restores files in the specified app directory
func GitRestore(appName string, files []string, staged bool) (GitRestoreResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitRestoreResult{}, err
	}

	if err := ValidateAppName(appName); err != nil {
		return GitRestoreResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitRestoreResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitRestoreResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitRestoreResult{
			IsRepo:  false,
			Success: false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	if len(files) == 0 {
		return GitRestoreResult{
			IsRepo:  true,
			Success: false,
			Message: "No files specified to restore",
		}, nil
	}

	// Validate file paths
	for _, file := range files {
		cleanPath := filepath.Clean(file)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			return GitRestoreResult{
				IsRepo:  true,
				Success: false,
				Message: fmt.Sprintf("Invalid file path: %s (must be relative to app directory)", file),
			}, nil
		}
	}

	// Build git restore command
	args := []string{"restore"}
	if staged {
		args = append(args, "--staged")
	}
	args = append(args, files...)

	// Run git restore
	restoreCmd := exec.Command("git", args...)
	restoreCmd.Dir = appPath
	output, err := restoreCmd.CombinedOutput()
	if err != nil {
		// Git restore doesn't exist in older versions, try checkout/reset
		if strings.Contains(string(output), "is not a git command") {
			if staged {
				// Use git reset for staged files
				args = []string{"reset", "HEAD"}
				args = append(args, files...)
			} else {
				// Use git checkout for working directory
				args = []string{"checkout", "--"}
				args = append(args, files...)
			}
			
			fallbackCmd := exec.Command("git", args...)
			fallbackCmd.Dir = appPath
			output, err = fallbackCmd.CombinedOutput()
			if err != nil {
				return GitRestoreResult{}, fmt.Errorf("git restore (fallback) failed: %w - %s", err, strings.TrimSpace(string(output)))
			}
		} else {
			return GitRestoreResult{}, fmt.Errorf("git restore failed: %w - %s", err, strings.TrimSpace(string(output)))
		}
	}

	return GitRestoreResult{
		IsRepo:        true,
		Success:       true,
		FilesRestored: files,
		Message:       "Files restored successfully",
	}, nil
}

// CLI
func GitRestoreCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("git_restore requires at least 2 arguments: app_name and files\nUsage: layered-code tool git_restore <app_name> <files...> [--staged]")
	}

	appName := args[0]
	var files []string
	staged := false

	// Parse arguments
	for i := 1; i < len(args); i++ {
		if args[i] == "--staged" {
			staged = true
		} else if !strings.HasPrefix(args[i], "-") {
			files = append(files, args[i])
		}
	}

	result, err := GitRestore(appName, files, staged)
	if err != nil {
		return fmt.Errorf("failed to restore files: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	if !result.Success {
		fmt.Printf("Failed: %s\n", result.Message)
		return nil
	}

	fmt.Println(result.Message)
	if len(result.FilesRestored) > 0 {
		fmt.Println("Restored files:")
		for _, file := range result.FilesRestored {
			fmt.Printf("  %s\n", file)
		}
	}

	return nil
}

// MCP
func GitRestoreMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string   `json:"app_name"`
		Files   []string `json:"files"`
		Staged  bool     `json:"staged"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	if len(args.Files) == 0 {
		return nil, fmt.Errorf("files are required")
	}

	result, err := GitRestore(args.AppName, args.Files, args.Staged)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}