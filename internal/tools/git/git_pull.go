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
type GitPullResult struct {
	Success bool   `json:"success"`
	IsRepo  bool   `json:"is_repo"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
	Updated bool   `json:"updated"`
}

// GitPull pulls changes from remote repository
func GitPull(appName string, remote string, branch string, rebase bool, lcmParams *LayeredChangeMemoryParams) (GitPullResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitPullResult{}, err
	}

	if err := ValidateAppName(appName); err != nil {
		return GitPullResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitPullResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitPullResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitPullResult{
			IsRepo:  false,
			Success: false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	// Default remote is "origin"
	if remote == "" {
		remote = "origin"
	}

	// Build git pull command
	args := []string{"pull"}
	
	if rebase {
		args = append(args, "--rebase")
	}
	
	args = append(args, remote)
	
	if branch != "" {
		args = append(args, branch)
	}

	// Run git pull
	pullCmd := exec.Command("git", args...)
	pullCmd.Dir = appPath
	output, err := pullCmd.CombinedOutput()
	
	outputStr := strings.TrimSpace(string(output))
	
	if err != nil {
		return GitPullResult{
			IsRepo:  true,
			Success: false,
			Message: "Pull failed",
			Output:  outputStr,
			Updated: false,
		}, nil
	}

	// Check if repository was updated
	updated := !strings.Contains(outputStr, "Already up to date")

	// Create LayeredChangeMemory entry if parameters provided and we pulled changes
	if lcmParams != nil && updated {
		// Get the current HEAD commit hash for the LCM entry
		hashCmd := exec.Command("git", "rev-parse", "HEAD")
		hashCmd.Dir = appPath
		hashOutput, err := hashCmd.Output()
		if err == nil {
			currentCommitHash := strings.TrimSpace(string(hashOutput))[:7] // Short hash
			message := fmt.Sprintf("Pulled changes from %s", remote)
			if branch != "" {
				message += fmt.Sprintf(" %s", branch)
			}
			if rebase {
				message += " (rebase)"
			}
			
			entry, err := GenerateLayeredChangeMemoryEntry(
				appPath,
				currentCommitHash,
				message,
				lcmParams.Summary,
				lcmParams.Considerations,
				lcmParams.FollowUp,
			)
			
			if err == nil {
				// Append to the LayeredChangeMemory log
				if err := AppendLayeredChangeMemoryEntry(appPath, entry); err != nil {
					// Log the error but don't fail the pull
					fmt.Fprintf(os.Stderr, "Warning: Failed to write LayeredChangeMemory entry: %v\n", err)
				}
			}
		}
	}

	return GitPullResult{
		IsRepo:  true,
		Success: true,
		Message: "Pull successful",
		Output:  outputStr,
		Updated: updated,
	}, nil
}

// CLI
func GitPullCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_pull requires at least 1 argument: app_name\nUsage: layered-code tool git_pull <app_name> [remote] [branch] [--rebase]")
	}

	appName := args[0]
	remote := ""
	branch := ""
	rebase := false

	// Parse arguments
	nonFlagArgs := []string{}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--rebase":
			rebase = true
		default:
			if !strings.HasPrefix(args[i], "-") {
				nonFlagArgs = append(nonFlagArgs, args[i])
			}
		}
	}

	// Assign non-flag arguments
	if len(nonFlagArgs) > 0 {
		remote = nonFlagArgs[0]
	}
	if len(nonFlagArgs) > 1 {
		branch = nonFlagArgs[1]
	}

	// For CLI, we don't support LayeredChangeMemory parameters
	result, err := GitPull(appName, remote, branch, rebase, nil)
	if err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	if result.Success {
		fmt.Println(result.Message)
		if result.Updated {
			fmt.Println("Repository updated with new changes")
		} else {
			fmt.Println("Already up to date")
		}
	} else {
		fmt.Printf("Pull failed: %s\n", result.Message)
	}

	if result.Output != "" {
		fmt.Println("\nOutput:")
		fmt.Println(result.Output)
	}

	return nil
}

// MCP
func GitPullMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName               string                    `json:"app_name"`
		Remote                string                    `json:"remote"`
		Branch                string                    `json:"branch"`
		Rebase                bool                      `json:"rebase"`
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

	result, err := GitPull(args.AppName, args.Remote, args.Branch, args.Rebase, args.LayeredChangeMemory)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}