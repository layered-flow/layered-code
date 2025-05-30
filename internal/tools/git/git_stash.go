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
type GitStashEntry struct {
	Index   string `json:"index"`
	Message string `json:"message"`
}

type GitStashResult struct {
	Success  bool            `json:"success"`
	Stashes  []GitStashEntry `json:"stashes"`
	IsRepo   bool            `json:"is_repo"`
	Message  string          `json:"message,omitempty"`
	Action   string          `json:"action,omitempty"`
}

// GitStash manages git stash in the specified app directory
func GitStash(appName string, action string, message string, lcmParams *LayeredChangeMemoryParams) (GitStashResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitStashResult{}, err
	}

	if err := ValidateAppName(appName); err != nil {
		return GitStashResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitStashResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitStashResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitStashResult{
			IsRepo:  false,
			Success: false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	result := GitStashResult{
		IsRepo: true,
		Action: action,
	}

	// Handle different stash actions
	switch action {
	case "push", "save":
		// Stash changes
		args := []string{"stash", "push"}
		if message != "" {
			args = append(args, "-m", message)
		}
		
		cmd := exec.Command("git", args...)
		cmd.Dir = appPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to stash: %s", strings.TrimSpace(string(output)))
		} else {
			result.Success = true
			result.Message = "Changes stashed successfully"
			
			// Create LayeredChangeMemory entry if parameters provided
			if lcmParams != nil {
				// Get the current HEAD commit hash for the LCM entry
				hashCmd := exec.Command("git", "rev-parse", "HEAD")
				hashCmd.Dir = appPath
				hashOutput, err := hashCmd.Output()
				if err == nil {
					currentCommitHash := strings.TrimSpace(string(hashOutput))[:7] // Short hash
					stashMessage := "Stashed changes"
					if message != "" {
						stashMessage = fmt.Sprintf("Stashed changes: %s", message)
					}
					
					entry, err := GenerateLayeredChangeMemoryEntry(
						appPath,
						currentCommitHash,
						stashMessage,
						lcmParams.Summary,
						lcmParams.Considerations,
						lcmParams.FollowUp,
					)
					
					if err == nil {
						// Append to the LayeredChangeMemory log
						if err := AppendLayeredChangeMemoryEntry(appPath, entry); err != nil {
							// Log the error but don't fail the stash
							fmt.Fprintf(os.Stderr, "Warning: Failed to write LayeredChangeMemory entry: %v\n", err)
						}
					}
				}
			}
		}

	case "pop":
		// Pop the latest stash
		cmd := exec.Command("git", "stash", "pop")
		cmd.Dir = appPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to pop stash: %s", strings.TrimSpace(string(output)))
		} else {
			result.Success = true
			result.Message = "Stash applied and removed"
			
			// Create LayeredChangeMemory entry if parameters provided
			if lcmParams != nil {
				// Get the current HEAD commit hash for the LCM entry
				hashCmd := exec.Command("git", "rev-parse", "HEAD")
				hashCmd.Dir = appPath
				hashOutput, err := hashCmd.Output()
				if err == nil {
					currentCommitHash := strings.TrimSpace(string(hashOutput))[:7] // Short hash
					
					entry, err := GenerateLayeredChangeMemoryEntry(
						appPath,
						currentCommitHash,
						"Popped stash",
						lcmParams.Summary,
						lcmParams.Considerations,
						lcmParams.FollowUp,
					)
					
					if err == nil {
						// Append to the LayeredChangeMemory log
						if err := AppendLayeredChangeMemoryEntry(appPath, entry); err != nil {
							// Log the error but don't fail the pop
							fmt.Fprintf(os.Stderr, "Warning: Failed to write LayeredChangeMemory entry: %v\n", err)
						}
					}
				}
			}
		}

	case "apply":
		// Apply the latest stash without removing it
		cmd := exec.Command("git", "stash", "apply")
		cmd.Dir = appPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to apply stash: %s", strings.TrimSpace(string(output)))
		} else {
			result.Success = true
			result.Message = "Stash applied"
		}

	case "drop":
		// Drop the latest stash
		cmd := exec.Command("git", "stash", "drop")
		cmd.Dir = appPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to drop stash: %s", strings.TrimSpace(string(output)))
		} else {
			result.Success = true
			result.Message = "Stash dropped"
		}

	case "list", "":
		// List stashes (default action)
		result.Success = true
		result.Action = "list"
	}

	// Always get the current stash list
	listCmd := exec.Command("git", "stash", "list")
	listCmd.Dir = appPath
	listOutput, err := listCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			
			// Parse stash entry (format: stash@{0}: message)
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) >= 2 {
				result.Stashes = append(result.Stashes, GitStashEntry{
					Index:   strings.TrimSpace(parts[0]),
					Message: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	return result, nil
}

// CLI
func GitStashCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_stash requires at least 1 argument: app_name\nUsage: layered-code tool git_stash <app_name> [push|pop|apply|drop|list] [-m \"message\"]")
	}

	appName := args[0]
	action := "list"
	message := ""

	// Parse arguments
	if len(args) > 1 {
		switch args[1] {
		case "push", "save", "pop", "apply", "drop", "list":
			action = args[1]
		}
	}

	// Look for message flag
	for i := 1; i < len(args)-1; i++ {
		if args[i] == "-m" || args[i] == "--message" {
			message = args[i+1]
			break
		}
	}

	// For CLI, we don't support LayeredChangeMemory parameters
	result, err := GitStash(appName, action, message, nil)
	if err != nil {
		return fmt.Errorf("failed to manage stash: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	// Show action result
	if action != "list" && action != "" {
		if result.Success {
			fmt.Println(result.Message)
		} else {
			fmt.Printf("Failed: %s\n", result.Message)
			return nil
		}
	}

	// Show stash list
	if len(result.Stashes) > 0 {
		fmt.Println("\nStash list:")
		for _, stash := range result.Stashes {
			fmt.Printf("  %s: %s\n", stash.Index, stash.Message)
		}
	} else if action == "list" || action == "" {
		fmt.Println("No stashes found")
	}

	return nil
}

// MCP
func GitStashMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName               string                    `json:"app_name"`
		Action                string                    `json:"action"`
		Message               string                    `json:"message"`
		LayeredChangeMemory   *LayeredChangeMemoryParams `json:"layered_change_memory,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	// Default to list if no action specified
	if args.Action == "" {
		args.Action = "list"
	}

	// Validate LayeredChangeMemory is provided for push/pop actions
	if (args.Action == "push" || args.Action == "save" || args.Action == "pop") && args.LayeredChangeMemory == nil {
		return nil, fmt.Errorf("layered_change_memory is required for %s action", args.Action)
	}

	// Validate LayeredChangeMemory fields when provided
	if args.LayeredChangeMemory != nil && args.LayeredChangeMemory.Summary == "" {
		return nil, fmt.Errorf("layered_change_memory.summary is required")
	}

	result, err := GitStash(args.AppName, args.Action, args.Message, args.LayeredChangeMemory)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}