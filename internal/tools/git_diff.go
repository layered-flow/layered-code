package tools

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
type GitDiffResult struct {
	Diff      string `json:"diff"`
	HasDiff   bool   `json:"has_diff"`
	IsRepo    bool   `json:"is_repo"`
	Message   string `json:"message,omitempty"`
	FileCount int    `json:"file_count"`
}

// GitDiff runs git diff command in the specified app directory
func GitDiff(appName string, staged bool, filePath string) (GitDiffResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitDiffResult{}, err
	}

	if err := validateAppName(appName); err != nil {
		return GitDiffResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitDiffResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := validateAppPath(appPath); err != nil {
		return GitDiffResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitDiffResult{
			IsRepo:  false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	// Build git diff command
	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}
	
	// Add unified context
	args = append(args, "-U3")
	
	// If a specific file path is provided, validate and add it
	if filePath != "" {
		// Validate the file path to ensure it's within the app directory
		cleanPath := filepath.Clean(filePath)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			return GitDiffResult{}, fmt.Errorf("invalid file path: must be relative to app directory")
		}
		args = append(args, "--", cleanPath)
	}

	// Run git diff
	diffCmd := exec.Command("git", args...)
	diffCmd.Dir = appPath
	output, err := diffCmd.Output()
	if err != nil {
		// Check if it's just an empty diff
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			output = []byte{}
		} else {
			return GitDiffResult{}, fmt.Errorf("failed to run git diff: %w", err)
		}
	}

	diff := string(output)
	hasDiff := len(strings.TrimSpace(diff)) > 0

	// Count the number of files in the diff
	fileCount := 0
	if hasDiff {
		lines := strings.Split(diff, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "diff --git") {
				fileCount++
			}
		}
	}

	return GitDiffResult{
		Diff:      diff,
		HasDiff:   hasDiff,
		IsRepo:    true,
		FileCount: fileCount,
	}, nil
}

// CLI
func GitDiffCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_diff requires at least 1 argument: app_name\nUsage: layered-code tool git_diff <app_name> [--staged] [file_path]")
	}

	appName := args[0]
	staged := false
	filePath := ""

	// Parse additional arguments
	for i := 1; i < len(args); i++ {
		if args[i] == "--staged" {
			staged = true
		} else if !strings.HasPrefix(args[i], "--") {
			filePath = args[i]
		}
	}

	result, err := GitDiff(appName, staged, filePath)
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	if !result.HasDiff {
		if staged {
			fmt.Println("No staged changes")
		} else {
			fmt.Println("No changes in working directory")
		}
		return nil
	}

	// Print the diff
	fmt.Print(result.Diff)

	return nil
}

// MCP
func GitDiffMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName  string `json:"app_name"`
		Staged   bool   `json:"staged"`
		FilePath string `json:"file_path"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitDiff(args.AppName, args.Staged, args.FilePath)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}