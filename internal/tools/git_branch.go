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
type GitBranchResult struct {
	Branches       []string `json:"branches"`
	CurrentBranch  string   `json:"current_branch"`
	IsRepo         bool     `json:"is_repo"`
	Message        string   `json:"message,omitempty"`
	CreateSuccess  bool     `json:"create_success,omitempty"`
	SwitchSuccess  bool     `json:"switch_success,omitempty"`
	DeleteSuccess  bool     `json:"delete_success,omitempty"`
}

// GitBranch manages git branches in the specified app directory
func GitBranch(appName string, createBranch string, switchBranch string, deleteBranch string, listAll bool) (GitBranchResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitBranchResult{}, err
	}

	if err := validateAppName(appName); err != nil {
		return GitBranchResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitBranchResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := validateAppPath(appPath); err != nil {
		return GitBranchResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitBranchResult{
			IsRepo:  false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	result := GitBranchResult{IsRepo: true}

	// Handle create branch
	if createBranch != "" {
		cmd := exec.Command("git", "branch", createBranch)
		cmd.Dir = appPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Message = fmt.Sprintf("Failed to create branch: %s", strings.TrimSpace(string(output)))
			result.CreateSuccess = false
		} else {
			result.CreateSuccess = true
		}
	}

	// Handle switch branch
	if switchBranch != "" {
		cmd := exec.Command("git", "checkout", switchBranch)
		cmd.Dir = appPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Message = fmt.Sprintf("Failed to switch branch: %s", strings.TrimSpace(string(output)))
			result.SwitchSuccess = false
		} else {
			result.SwitchSuccess = true
		}
	}

	// Handle delete branch
	if deleteBranch != "" {
		cmd := exec.Command("git", "branch", "-d", deleteBranch)
		cmd.Dir = appPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Try force delete if regular delete fails
			cmd = exec.Command("git", "branch", "-D", deleteBranch)
			cmd.Dir = appPath
			output, err = cmd.CombinedOutput()
			if err != nil {
				result.Message = fmt.Sprintf("Failed to delete branch: %s", strings.TrimSpace(string(output)))
				result.DeleteSuccess = false
			} else {
				result.DeleteSuccess = true
			}
		} else {
			result.DeleteSuccess = true
		}
	}

	// Get current branch
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = appPath
	branchOutput, err := branchCmd.Output()
	if err != nil {
		// Try alternative method for detached HEAD
		branchCmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branchCmd.Dir = appPath
		branchOutput, err = branchCmd.Output()
		if err != nil {
			result.CurrentBranch = "unknown"
		} else {
			result.CurrentBranch = strings.TrimSpace(string(branchOutput))
		}
	} else {
		result.CurrentBranch = strings.TrimSpace(string(branchOutput))
	}

	// List branches
	args := []string{"branch"}
	if listAll {
		args = append(args, "-a")
	}

	listCmd := exec.Command("git", args...)
	listCmd.Dir = appPath
	output, err := listCmd.Output()
	if err != nil {
		return result, nil
	}

	// Parse branch list
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove the * indicator for current branch
		if strings.HasPrefix(line, "* ") {
			line = line[2:]
		}

		// Clean up remote branch names
		if strings.HasPrefix(line, "remotes/") {
			line = strings.TrimPrefix(line, "remotes/")
		}

		result.Branches = append(result.Branches, line)
	}

	return result, nil
}

// CLI
func GitBranchCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_branch requires at least 1 argument: app_name\nUsage: layered-code tool git_branch <app_name> [-c <branch>] [-s <branch>] [-d <branch>] [-a]")
	}

	appName := args[0]
	createBranch := ""
	switchBranch := ""
	deleteBranch := ""
	listAll := false

	// Parse arguments
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-c", "--create":
			if i+1 < len(args) {
				createBranch = args[i+1]
				i++
			}
		case "-s", "--switch":
			if i+1 < len(args) {
				switchBranch = args[i+1]
				i++
			}
		case "-d", "--delete":
			if i+1 < len(args) {
				deleteBranch = args[i+1]
				i++
			}
		case "-a", "--all":
			listAll = true
		}
	}

	result, err := GitBranch(appName, createBranch, switchBranch, deleteBranch, listAll)
	if err != nil {
		return fmt.Errorf("failed to manage branches: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	// Show operation results
	if createBranch != "" {
		if result.CreateSuccess {
			fmt.Printf("Created branch: %s\n", createBranch)
		} else {
			fmt.Printf("Failed to create branch: %s\n", result.Message)
		}
	}

	if switchBranch != "" {
		if result.SwitchSuccess {
			fmt.Printf("Switched to branch: %s\n", switchBranch)
		} else {
			fmt.Printf("Failed to switch branch: %s\n", result.Message)
		}
	}

	if deleteBranch != "" {
		if result.DeleteSuccess {
			fmt.Printf("Deleted branch: %s\n", deleteBranch)
		} else {
			fmt.Printf("Failed to delete branch: %s\n", result.Message)
		}
	}

	// Show branch list
	if createBranch == "" && switchBranch == "" && deleteBranch == "" {
		fmt.Printf("Current branch: %s\n\n", result.CurrentBranch)
		fmt.Println("Branches:")
		for _, branch := range result.Branches {
			if branch == result.CurrentBranch {
				fmt.Printf("* %s\n", branch)
			} else {
				fmt.Printf("  %s\n", branch)
			}
		}
	}

	return nil
}

// MCP
func GitBranchMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName      string `json:"app_name"`
		CreateBranch string `json:"create_branch"`
		SwitchBranch string `json:"switch_branch"`
		DeleteBranch string `json:"delete_branch"`
		ListAll      bool   `json:"list_all"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitBranch(args.AppName, args.CreateBranch, args.SwitchBranch, args.DeleteBranch, args.ListAll)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}