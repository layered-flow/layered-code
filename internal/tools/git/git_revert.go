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

// Revert creates a revert commit that undoes changes from a previous commit
func Revert(appName, commitHash string, noCommit bool) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("app_name is required")
	}
	if commitHash == "" {
		return "", fmt.Errorf("commit_hash is required")
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return "", fmt.Errorf("failed to get apps directory: %w", err)
	}
	repoPath := filepath.Join(appsDir, appName)
	
	// Build command
	args := []string{"revert", "--no-edit"}
	if noCommit {
		args = append(args, "--no-commit")
	}
	args = append(args, commitHash)
	
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git revert failed: %w\nOutput: %s", err, string(output))
	}

	var result string
	if noCommit {
		// Show what was reverted but not committed
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusCmd.Dir = repoPath
		statusOutput, _ := statusCmd.CombinedOutput()
		
		result = fmt.Sprintf("Successfully reverted commit %s (changes staged but not committed)", commitHash)
		if len(statusOutput) > 0 {
			result += fmt.Sprintf("\n\nStaged changes:\n%s", string(statusOutput))
		}
	} else {
		// Show the new revert commit
		logCmd := exec.Command("git", "log", "--oneline", "-1")
		logCmd.Dir = repoPath
		logOutput, _ := logCmd.CombinedOutput()
		
		result = fmt.Sprintf("Successfully created revert commit for %s", commitHash)
		if len(logOutput) > 0 {
			result += fmt.Sprintf("\n\nNew commit: %s", strings.TrimSpace(string(logOutput)))
		}
		
		// Also show what was reverted
		diffCmd := exec.Command("git", "diff", "HEAD~1", "HEAD", "--stat")
		diffCmd.Dir = repoPath
		if diffOutput, err := diffCmd.CombinedOutput(); err == nil && len(diffOutput) > 0 {
			result += fmt.Sprintf("\n\nChanges reverted:\n%s", string(diffOutput))
		}
	}

	return result, nil
}

// CLI
func GitRevertCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("git_revert requires at least 2 arguments: app_name commit_hash [--no-commit]\nUsage: layered-code tool git_revert <app_name> <commit_hash> [--no-commit]")
	}

	appName := args[0]
	commitHash := args[1]
	
	noCommit := false
	if len(args) > 2 && args[2] == "--no-commit" {
		noCommit = true
	}

	result, err := Revert(appName, commitHash, noCommit)
	if err != nil {
		return fmt.Errorf("failed to revert: %w", err)
	}

	fmt.Println(result)
	return nil
}

// MCP
func GitRevertMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName    string `json:"app_name"`
		CommitHash string `json:"commit_hash"`
		NoCommit   bool   `json:"no_commit,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := Revert(args.AppName, args.CommitHash, args.NoCommit)
	if err != nil {
		return nil, err
	}

	// Return structured result
	type Result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	response := Result{
		Success: true,
		Message: result,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(responseBytes),
			},
		},
	}, nil
}