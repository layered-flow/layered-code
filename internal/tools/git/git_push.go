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
type GitPushResult struct {
	Success     bool   `json:"success"`
	IsRepo      bool   `json:"is_repo"`
	Message     string `json:"message"`
	Output      string `json:"output,omitempty"`
	ErrorOutput string `json:"error_output,omitempty"`
}

// GitPush pushes commits to remote repository
func GitPush(appName string, remote string, branch string, setUpstream bool, force bool) (GitPushResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitPushResult{}, err
	}

	if err := helpers.ValidateAppName(appName); err != nil {
		return GitPushResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitPushResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitPushResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitPushResult{
			IsRepo:  false,
			Success: false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	// Default remote is "origin"
	if remote == "" {
		remote = "origin"
	}

	// Build git push command
	args := []string{"push"}
	
	if setUpstream {
		args = append(args, "-u")
	}
	
	if force {
		args = append(args, "--force")
	}
	
	args = append(args, remote)
	
	if branch != "" {
		args = append(args, branch)
	}

	// Run git push
	pushCmd := exec.Command("git", args...)
	pushCmd.Dir = appPath
	var outBuf, errBuf bytes.Buffer
	pushCmd.Stdout = &outBuf
	pushCmd.Stderr = &errBuf
	err = pushCmd.Run()
	
	outputStr := strings.TrimSpace(outBuf.String())
	errorStr := strings.TrimSpace(errBuf.String())
	
	if err != nil {
		return GitPushResult{
			IsRepo:      true,
			Success:     false,
			Message:     "Git push failed",
			Output:      outputStr,
			ErrorOutput: errorStr,
		}, fmt.Errorf("git push failed: %w - %s", err, errorStr)
	}

	return GitPushResult{
		IsRepo:  true,
		Success: true,
		Message: "Push successful",
		Output:  outputStr,
	}, nil
}

// CLI
func GitPushCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_push requires at least 1 argument: app_name\nUsage: layered-code tool git_push <app_name> [remote] [branch] [-u|--set-upstream] [--force]")
	}

	appName := args[0]
	remote := ""
	branch := ""
	setUpstream := false
	force := false

	// Parse arguments
	nonFlagArgs := []string{}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-u", "--set-upstream":
			setUpstream = true
		case "--force", "-f":
			force = true
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

	result, err := GitPush(appName, remote, branch, setUpstream, force)
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	if result.Success {
		fmt.Println(result.Message)
	} else {
		fmt.Printf("Push failed: %s\n", result.Message)
	}

	if result.Output != "" {
		fmt.Println(result.Output)
	}

	return nil
}

// MCP
func GitPushMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName     string `json:"app_name"`
		Remote      string `json:"remote"`
		Branch      string `json:"branch"`
		SetUpstream bool   `json:"set_upstream"`
		Force       bool   `json:"force"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitPush(args.AppName, args.Remote, args.Branch, args.SetUpstream, args.Force)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}