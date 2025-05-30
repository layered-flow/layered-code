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

// Checkout switches branches or restores working tree files
func Checkout(appName, target string, isNewBranch bool, files []string, lcmParams *LayeredChangeMemoryParams) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("app_name is required")
	}
	if target == "" && len(files) == 0 {
		return "", fmt.Errorf("either target branch/commit or files must be specified")
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return "", fmt.Errorf("failed to get apps directory: %w", err)
	}
	repoPath := filepath.Join(appsDir, appName)
	
	var args []string
	var operation string
	
	if len(files) > 0 {
		// Checkout specific files
		args = []string{"checkout"}
		if target != "" {
			// Checkout files from specific commit/branch
			args = append(args, target, "--")
			operation = fmt.Sprintf("files from %s", target)
		} else {
			// Checkout files from HEAD
			args = append(args, "HEAD", "--")
			operation = "files from HEAD"
		}
		args = append(args, files...)
	} else {
		// Checkout branch or commit
		args = []string{"checkout"}
		if isNewBranch {
			args = append(args, "-b")
			operation = fmt.Sprintf("new branch '%s'", target)
		} else {
			operation = fmt.Sprintf("branch/commit '%s'", target)
		}
		args = append(args, target)
	}
	
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git checkout failed: %w\nOutput: %s", err, string(output))
	}

	result := fmt.Sprintf("Successfully checked out %s", operation)
	
	// Get current branch/HEAD info
	if len(files) == 0 {
		// Show current branch after checkout
		branchCmd := exec.Command("git", "branch", "--show-current")
		branchCmd.Dir = repoPath
		if branchOutput, err := branchCmd.CombinedOutput(); err == nil {
			branch := strings.TrimSpace(string(branchOutput))
			if branch != "" {
				result += fmt.Sprintf("\n\nCurrent branch: %s", branch)
			} else {
				// Detached HEAD state
				headCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
				headCmd.Dir = repoPath
				if headOutput, err := headCmd.CombinedOutput(); err == nil {
					result += fmt.Sprintf("\n\nHEAD is now at %s (detached)", strings.TrimSpace(string(headOutput)))
				}
			}
		}
		
		// Show recent commits
		logCmd := exec.Command("git", "log", "--oneline", "-5")
		logCmd.Dir = repoPath
		if logOutput, err := logCmd.CombinedOutput(); err == nil {
			result += fmt.Sprintf("\n\nRecent commits:\n%s", string(logOutput))
		}
	} else {
		// Show status of checked out files
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusCmd.Dir = repoPath
		if statusOutput, err := statusCmd.CombinedOutput(); err == nil && len(statusOutput) > 0 {
			result += fmt.Sprintf("\n\nFile status:\n%s", string(statusOutput))
		}
		
		// Create LayeredChangeMemory entry if parameters provided and we're restoring files
		if lcmParams != nil {
			// Get the current HEAD commit hash for the LCM entry
			hashCmd := exec.Command("git", "rev-parse", "HEAD")
			hashCmd.Dir = repoPath
			hashOutput, err := hashCmd.Output()
			if err == nil {
				currentCommitHash := strings.TrimSpace(string(hashOutput))[:7] // Short hash
				var message string
				if target != "" && target != "HEAD" {
					message = fmt.Sprintf("Restored %d file(s) from %s", len(files), target)
				} else {
					message = fmt.Sprintf("Restored %d file(s) from HEAD", len(files))
				}
				
				entry, err := GenerateLayeredChangeMemoryEntry(
					repoPath,
					currentCommitHash,
					message,
					lcmParams.Summary,
					lcmParams.Considerations,
					lcmParams.FollowUp,
				)
				
				if err == nil {
					// Append to the LayeredChangeMemory log
					if err := AppendLayeredChangeMemoryEntry(repoPath, entry); err != nil {
						// Log the error but don't fail the checkout
						fmt.Fprintf(os.Stderr, "Warning: Failed to write LayeredChangeMemory entry: %v\n", err)
					}
				}
			}
		}
	}

	return result, nil
}

// CLI
func GitCheckoutCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("git_checkout requires at least 2 arguments\nUsage:\n  layered-code tool git_checkout <app_name> <branch/commit> [-b]\n  layered-code tool git_checkout <app_name> --files <file1> [file2 ...]\n  layered-code tool git_checkout <app_name> <commit> --files <file1> [file2 ...]")
	}

	appName := args[0]
	
	// Check if we're checking out files
	filesIndex := -1
	for i, arg := range args[1:] {
		if arg == "--files" {
			filesIndex = i + 1
			break
		}
	}
	
	if filesIndex > 0 {
		// Checking out files
		var target string
		var files []string
		
		if filesIndex == 1 {
			// git_checkout app --files file1 file2
			if len(args) < 3 {
				return fmt.Errorf("--files requires at least one file")
			}
			files = args[filesIndex+1:]
		} else {
			// git_checkout app commit --files file1 file2
			target = args[1]
			if len(args) < filesIndex+2 {
				return fmt.Errorf("--files requires at least one file")
			}
			files = args[filesIndex+1:]
		}
		
		// For CLI, we don't support LayeredChangeMemory parameters
		result, err := Checkout(appName, target, false, files, nil)
		if err != nil {
			return fmt.Errorf("failed to checkout files: %w", err)
		}
		fmt.Println(result)
	} else {
		// Checking out branch/commit
		target := args[1]
		isNewBranch := false
		
		if len(args) > 2 && args[2] == "-b" {
			isNewBranch = true
		}
		
		// For CLI, we don't support LayeredChangeMemory parameters
		result, err := Checkout(appName, target, isNewBranch, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to checkout: %w", err)
		}
		fmt.Println(result)
	}

	return nil
}

// MCP
func GitCheckoutMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName               string                    `json:"app_name"`
		Target                string                    `json:"target,omitempty"`
		IsNewBranch           bool                      `json:"is_new_branch,omitempty"`
		Files                 []string                  `json:"files,omitempty"`
		LayeredChangeMemory   *LayeredChangeMemoryParams `json:"layered_change_memory,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate that either target or files is specified
	if args.Target == "" && len(args.Files) == 0 {
		return nil, fmt.Errorf("either target branch/commit or files must be specified")
	}

	// Validate LayeredChangeMemory is provided when restoring files
	if len(args.Files) > 0 && args.LayeredChangeMemory == nil {
		return nil, fmt.Errorf("layered_change_memory is required when restoring files")
	}

	// Validate LayeredChangeMemory fields when provided
	if args.LayeredChangeMemory != nil && args.LayeredChangeMemory.Summary == "" {
		return nil, fmt.Errorf("layered_change_memory.summary is required")
	}

	result, err := Checkout(args.AppName, args.Target, args.IsNewBranch, args.Files, args.LayeredChangeMemory)
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