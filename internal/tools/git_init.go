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
type GitInitResult struct {
	Success       bool   `json:"success"`
	AlreadyExists bool   `json:"already_exists"`
	Message       string `json:"message"`
	AppPath       string `json:"app_path"`
}

// GitInit initializes a git repository in the specified app directory
func GitInit(appName string, bare bool) (GitInitResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitInitResult{}, err
	}

	if err := validateAppName(appName); err != nil {
		return GitInitResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitInitResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	
	// Create app directory if it doesn't exist
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		if err := os.MkdirAll(appPath, 0755); err != nil {
			return GitInitResult{}, fmt.Errorf("failed to create app directory: %w", err)
		}
	}

	// Check if it's already a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return GitInitResult{
			Success:       false,
			AlreadyExists: true,
			Message:       fmt.Sprintf("'%s' is already a git repository", appName),
			AppPath:       appPath,
		}, nil
	}

	// Build git init command
	args := []string{"init"}
	if bare {
		args = append(args, "--bare")
	}

	// Run git init
	initCmd := exec.Command("git", args...)
	initCmd.Dir = appPath
	output, err := initCmd.CombinedOutput()
	if err != nil {
		return GitInitResult{
			Success: false,
			Message: fmt.Sprintf("Failed to initialize repository: %s", strings.TrimSpace(string(output))),
			AppPath: appPath,
		}, nil
	}

	// Configure default branch name to "main"
	configCmd := exec.Command("git", "config", "init.defaultBranch", "main")
	configCmd.Dir = appPath
	configCmd.Run() // Ignore errors for older git versions

	return GitInitResult{
		Success: true,
		Message: fmt.Sprintf("Initialized git repository in '%s'", appName),
		AppPath: appPath,
	}, nil
}

// CLI
func GitInitCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_init requires at least 1 argument: app_name\nUsage: layered-code tool git_init <app_name> [--bare]")
	}

	appName := args[0]
	bare := false

	// Parse arguments
	for i := 1; i < len(args); i++ {
		if args[i] == "--bare" {
			bare = true
		}
	}

	result, err := GitInit(appName, bare)
	if err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	if result.AlreadyExists {
		fmt.Println(result.Message)
		return nil
	}

	if result.Success {
		fmt.Println(result.Message)
		fmt.Printf("Path: %s\n", result.AppPath)
	} else {
		fmt.Printf("Failed: %s\n", result.Message)
	}

	return nil
}

// MCP
func GitInitMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
		Bare    bool   `json:"bare"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitInit(args.AppName, args.Bare)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}