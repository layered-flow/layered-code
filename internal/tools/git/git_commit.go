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
type GitCommitResult struct {
	Success     bool   `json:"success"`
	CommitHash  string `json:"commit_hash"`
	Message     string `json:"message"`
	IsRepo      bool   `json:"is_repo"`
	Error       string `json:"error,omitempty"`
	ErrorOutput string `json:"error_output,omitempty"`
}

// GitCommit creates a git commit in the specified app directory
func GitCommit(appName string, message string, amend bool) (GitCommitResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitCommitResult{}, err
	}

	if err := helpers.ValidateAppName(appName); err != nil {
		return GitCommitResult{}, err
	}

	if message == "" && !amend {
		return GitCommitResult{}, fmt.Errorf("commit message is required (unless using --amend)")
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitCommitResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitCommitResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitCommitResult{
			IsRepo:  false,
			Success: false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	// Check if there are staged changes
	statusCmd := exec.Command("git", "diff", "--staged", "--quiet")
	statusCmd.Dir = appPath
	err = statusCmd.Run()
	hasStaged := err != nil // git diff --staged --quiet returns non-zero if there are staged changes

	if !hasStaged && !amend {
		return GitCommitResult{
			IsRepo:  true,
			Success: false,
			Message: "No staged changes to commit",
		}, nil
	}

	// Build commit command
	args := []string{"commit"}
	if amend {
		args = append(args, "--amend")
		if message != "" {
			args = append(args, "-m", message)
		} else {
			args = append(args, "--no-edit")
		}
	} else {
		args = append(args, "-m", message)
	}

	// Run git commit
	commitCmd := exec.Command("git", args...)
	commitCmd.Dir = appPath
	var outBuf, errBuf bytes.Buffer
	commitCmd.Stdout = &outBuf
	commitCmd.Stderr = &errBuf
	err = commitCmd.Run()
	if err != nil {
		return GitCommitResult{
			IsRepo:      true,
			Success:     false,
			Message:     "Git commit failed",
			Error:       err.Error(),
			ErrorOutput: errBuf.String(),
		}, fmt.Errorf("git commit failed: %w - %s", err, strings.TrimSpace(errBuf.String()))
	}

	// Get the commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = appPath
	var hashOutBuf, hashErrBuf bytes.Buffer
	hashCmd.Stdout = &hashOutBuf
	hashCmd.Stderr = &hashErrBuf
	err = hashCmd.Run()
	if err != nil {
		return GitCommitResult{
			IsRepo:      true,
			Success:     false,
			Message:     "Failed to get commit hash",
			Error:       err.Error(),
			ErrorOutput: hashErrBuf.String(),
		}, fmt.Errorf("failed to get commit hash: %w", err)
	}

	commitHash := strings.TrimSpace(hashOutBuf.String())[:7] // Short hash

	return GitCommitResult{
		IsRepo:     true,
		Success:    true,
		CommitHash: commitHash,
		Message:    "Commit created successfully",
	}, nil
}

// CLI
func GitCommitCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_commit requires at least 1 argument: app_name\nUsage: layered-code tool git_commit <app_name> [-m \"message\"] [--amend]")
	}

	appName := args[0]
	message := ""
	amend := false

	// Parse arguments
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++ // Skip the message value
			} else {
				return fmt.Errorf("-m flag requires a message")
			}
		case "--amend":
			amend = true
		}
	}

	result, err := GitCommit(appName, message, amend)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	if !result.Success {
		fmt.Printf("Commit failed: %s\n", result.Message)
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
		return nil
	}

	fmt.Printf("Commit created successfully: %s\n", result.CommitHash)
	return nil
}

// MCP
func GitCommitMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
		Message string `json:"message"`
		Amend   bool   `json:"amend"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitCommit(args.AppName, args.Message, args.Amend)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}