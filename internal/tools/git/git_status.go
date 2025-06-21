package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/helpers"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type GitStatusResult struct {
	Branch      string   `json:"branch"`
	Staged      []string `json:"staged"`
	Modified    []string `json:"modified"`
	Untracked   []string `json:"untracked"`
	IsRepo      bool     `json:"is_repo"`
	Message     string   `json:"message,omitempty"`
	ErrorOutput string   `json:"error_output,omitempty"`
}

// GitStatus runs git status command in the specified app directory
func GitStatus(appName string) (GitStatusResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitStatusResult{}, err
	}

	if err := helpers.ValidateAppName(appName); err != nil {
		return GitStatusResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitStatusResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitStatusResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitStatusResult{
			IsRepo:  false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	// Get current branch
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = appPath
	branchOutput, branchErr := branchCmd.CombinedOutput()
	if branchErr != nil {
		// Try alternative method for detached HEAD
		branchCmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branchCmd.Dir = appPath
		branchOutput, branchErr = branchCmd.CombinedOutput()
		if branchErr != nil {
			return GitStatusResult{}, fmt.Errorf("failed to get current branch: %w", branchErr)
		}
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get git status
	statusCmd := exec.Command("git", "status", "--porcelain=v1")
	statusCmd.Dir = appPath
	var statusOut, statusErr bytes.Buffer
	statusCmd.Stdout = &statusOut
	statusCmd.Stderr = &statusErr
	err = statusCmd.Run()
	if err != nil {
		return GitStatusResult{}, fmt.Errorf("failed to run git status: %w", err)
	}

	result := GitStatusResult{
		Branch:      branch,
		IsRepo:      true,
		Staged:      []string{},
		Modified:    []string{},
		Untracked:   []string{},
		ErrorOutput: statusErr.String(),
	}

	// Parse git status output
	lines := strings.Split(statusOut.String(), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		filename := strings.TrimSpace(line[3:])

		switch {
		case statusCode[0] != ' ' && statusCode[0] != '?':
			// Staged file
			result.Staged = append(result.Staged, filename)
		case statusCode[1] == 'M' || statusCode[1] == 'D':
			// Modified or deleted in working tree
			result.Modified = append(result.Modified, filename)
		case statusCode == "??":
			// Untracked file
			result.Untracked = append(result.Untracked, filename)
		}
	}

	return result, nil
}

// CLI
func GitStatusCli() error {
	args := os.Args[3:]

	if len(args) != 1 {
		return fmt.Errorf("git_status requires exactly 1 argument: app_name\nUsage: layered-code tool git_status <app_name>")
	}

	appName := args[0]
	result, err := GitStatus(appName)
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	fmt.Printf("Git Status for '%s':\n", appName)
	fmt.Printf("Branch: %s\n", result.Branch)

	if len(result.Staged) > 0 {
		fmt.Println("\nStaged changes:")
		for _, file := range result.Staged {
			fmt.Printf("  + %s\n", file)
		}
	}

	if len(result.Modified) > 0 {
		fmt.Println("\nModified files:")
		for _, file := range result.Modified {
			fmt.Printf("  M %s\n", file)
		}
	}

	if len(result.Untracked) > 0 {
		fmt.Println("\nUntracked files:")
		for _, file := range result.Untracked {
			fmt.Printf("  ? %s\n", file)
		}
	}

	if len(result.Staged) == 0 && len(result.Modified) == 0 && len(result.Untracked) == 0 {
		fmt.Println("\nWorking tree clean")
	}

	return nil
}

// MCP
func GitStatusMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitStatus(args.AppName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}