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
type GitCommitResult struct {
	Success    bool   `json:"success"`
	CommitHash string `json:"commit_hash"`
	Message    string `json:"message"`
	IsRepo     bool   `json:"is_repo"`
	Error      string `json:"error,omitempty"`
}

// LayeredChangeMemoryParams contains the parameters for creating a LayeredChangeMemory entry
type LayeredChangeMemoryParams struct {
	Summary        string   `json:"summary"`
	Considerations []string `json:"considerations"`
	FollowUp       string   `json:"follow_up"`
}

// GitCommit creates a git commit in the specified app directory
func GitCommit(appName string, message string, amend bool, lcmParams *LayeredChangeMemoryParams) (GitCommitResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitCommitResult{}, err
	}

	if err := ValidateAppName(appName); err != nil {
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
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		return GitCommitResult{
			IsRepo:  true,
			Success: false,
			Error:   strings.TrimSpace(string(output)),
			Message: "Commit failed",
		}, nil
	}

	// Get the commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = appPath
	hashOutput, err := hashCmd.Output()
	if err != nil {
		return GitCommitResult{
			IsRepo:  true,
			Success: true,
			Message: "Commit created but failed to get hash",
		}, nil
	}

	commitHash := strings.TrimSpace(string(hashOutput))[:7] // Short hash

	// Create LayeredChangeMemory entry if parameters provided
	if lcmParams != nil {
		entry, err := GenerateLayeredChangeMemoryEntry(
			appPath,
			commitHash,
			message,
			lcmParams.Summary,
			lcmParams.Considerations,
			lcmParams.FollowUp,
		)
		
		if err == nil {
			// Append to the LayeredChangeMemory log
			if err := AppendLayeredChangeMemoryEntry(appPath, entry); err != nil {
				// Log the error but don't fail the commit
				fmt.Fprintf(os.Stderr, "Warning: Failed to write LayeredChangeMemory entry: %v\n", err)
			}
		}
	}

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

	// For CLI, we don't support LayeredChangeMemory parameters
	result, err := GitCommit(appName, message, amend, nil)
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
		AppName               string                    `json:"app_name"`
		Message               string                    `json:"message"`
		Amend                 bool                      `json:"amend"`
		LayeredChangeMemory   *LayeredChangeMemoryParams `json:"layered_change_memory,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	// Validate LayeredChangeMemory is provided
	if args.LayeredChangeMemory == nil {
		return nil, fmt.Errorf("layered_change_memory is required")
	}

	// Validate LayeredChangeMemory fields
	if args.LayeredChangeMemory.Summary == "" {
		return nil, fmt.Errorf("layered_change_memory.summary is required")
	}

	result, err := GitCommit(args.AppName, args.Message, args.Amend, args.LayeredChangeMemory)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}