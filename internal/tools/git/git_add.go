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
type GitAddResult struct {
	Success     bool     `json:"success"`
	FilesAdded  []string `json:"files_added"`
	IsRepo      bool     `json:"is_repo"`
	Message     string   `json:"message,omitempty"`
}

// GitAdd stages files in the specified app directory
func GitAdd(appName string, files []string, all bool) (GitAddResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitAddResult{}, err
	}

	if err := ValidateAppName(appName); err != nil {
		return GitAddResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitAddResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitAddResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitAddResult{
			IsRepo:  false,
			Success: false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	// Build git add command
	args := []string{"add"}
	
	if all {
		args = append(args, "-A")
	} else if len(files) == 0 {
		return GitAddResult{
			IsRepo:  true,
			Success: false,
			Message: "No files specified to add",
		}, nil
	} else {
		// Validate file paths
		for _, file := range files {
			cleanPath := filepath.Clean(file)
			if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
				return GitAddResult{
					IsRepo:  true,
					Success: false,
					Message: fmt.Sprintf("Invalid file path: %s (must be relative to app directory)", file),
				}, nil
			}
		}
		args = append(args, files...)
	}

	// Run git add
	addCmd := exec.Command("git", args...)
	addCmd.Dir = appPath
	output, err := addCmd.CombinedOutput()
	if err != nil {
		return GitAddResult{
			IsRepo:  true,
			Success: false,
			Message: fmt.Sprintf("Failed to add files: %s", strings.TrimSpace(string(output))),
		}, nil
	}

	// Get list of staged files
	statusCmd := exec.Command("git", "diff", "--name-only", "--cached")
	statusCmd.Dir = appPath
	statusOutput, err := statusCmd.Output()
	
	var filesAdded []string
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
		for _, line := range lines {
			if line != "" {
				filesAdded = append(filesAdded, line)
			}
		}
	}

	return GitAddResult{
		IsRepo:     true,
		Success:    true,
		FilesAdded: filesAdded,
		Message:    "Files staged successfully",
	}, nil
}

// CLI
func GitAddCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_add requires at least 1 argument: app_name\nUsage: layered-code tool git_add <app_name> [files...] [-A|--all]")
	}

	appName := args[0]
	var files []string
	all := false

	// Parse arguments
	for i := 1; i < len(args); i++ {
		if args[i] == "-A" || args[i] == "--all" {
			all = true
		} else if !strings.HasPrefix(args[i], "-") {
			files = append(files, args[i])
		}
	}

	result, err := GitAdd(appName, files, all)
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
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
	if len(result.FilesAdded) > 0 {
		fmt.Println("Staged files:")
		for _, file := range result.FilesAdded {
			fmt.Printf("  %s\n", file)
		}
	}

	return nil
}

// MCP
func GitAddMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string   `json:"app_name"`
		Files   []string `json:"files"`
		All     bool     `json:"all"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitAdd(args.AppName, args.Files, args.All)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}