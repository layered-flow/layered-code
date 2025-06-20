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

// Checkout switches branches or restores working tree files
func Checkout(appName, target string, isNewBranch bool, files []string) (string, error) {
	if err := helpers.ValidateAppName(appName); err != nil {
		return "", err
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
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git checkout failed: %w\nError output: %s", err, errBuf.String())
	}

	result := fmt.Sprintf("Successfully checked out %s", operation)
	
	// Get current branch/HEAD info
	if len(files) == 0 {
		// Show current branch after checkout
		branchCmd := exec.Command("git", "branch", "--show-current")
		branchCmd.Dir = repoPath
		var branchOutBuf, branchErrBuf bytes.Buffer
		branchCmd.Stdout = &branchOutBuf
		branchCmd.Stderr = &branchErrBuf
		if err := branchCmd.Run(); err == nil {
			branch := strings.TrimSpace(branchOutBuf.String())
			if branch != "" {
				result += fmt.Sprintf("\n\nCurrent branch: %s", branch)
			} else {
				// Detached HEAD state
				headCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
				headCmd.Dir = repoPath
				var headOutBuf, headErrBuf bytes.Buffer
				headCmd.Stdout = &headOutBuf
				headCmd.Stderr = &headErrBuf
				if err := headCmd.Run(); err == nil {
					result += fmt.Sprintf("\n\nHEAD is now at %s (detached)", strings.TrimSpace(headOutBuf.String()))
				}
			}
		}
		
		// Show recent commits
		logCmd := exec.Command("git", "log", "--oneline", "-5")
		logCmd.Dir = repoPath
		var logOutBuf, logErrBuf bytes.Buffer
		logCmd.Stdout = &logOutBuf
		logCmd.Stderr = &logErrBuf
		if err := logCmd.Run(); err == nil {
			result += fmt.Sprintf("\n\nRecent commits:\n%s", logOutBuf.String())
		}
	} else {
		// Show status of checked out files
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusCmd.Dir = repoPath
		var statusOutBuf, statusErrBuf bytes.Buffer
		statusCmd.Stdout = &statusOutBuf
		statusCmd.Stderr = &statusErrBuf
		if err := statusCmd.Run(); err == nil && statusOutBuf.Len() > 0 {
			result += fmt.Sprintf("\n\nFile status:\n%s", statusOutBuf.String())
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
		
		result, err := Checkout(appName, target, false, files)
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
		
		result, err := Checkout(appName, target, isNewBranch, nil)
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
		AppName     string   `json:"app_name"`
		Target      string   `json:"target,omitempty"`
		IsNewBranch bool     `json:"is_new_branch,omitempty"`
		Files       []string `json:"files,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate that either target or files is specified
	if args.Target == "" && len(args.Files) == 0 {
		return nil, fmt.Errorf("either target branch/commit or files must be specified")
	}

	result, err := Checkout(args.AppName, args.Target, args.IsNewBranch, args.Files)
	if err != nil {
		return nil, err
	}

	// Return structured result
	type Result struct {
		Success     bool   `json:"success"`
		Message     string `json:"message"`
		ErrorOutput string `json:"error_output,omitempty"`
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