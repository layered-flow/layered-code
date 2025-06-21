package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/helpers"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type GitLogEntry struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

type GitLogResult struct {
	Commits     []GitLogEntry `json:"commits"`
	IsRepo      bool          `json:"is_repo"`
	Message     string        `json:"message,omitempty"`
	ErrorOutput string        `json:"error_output,omitempty"`
}

// GitLog retrieves git log for the specified app directory
func GitLog(appName string, limit int, oneline bool) (GitLogResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitLogResult{}, err
	}

	if err := helpers.ValidateAppName(appName); err != nil {
		return GitLogResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitLogResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitLogResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitLogResult{
			IsRepo:  false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	// Build git log command
	args := []string{"log"}
	
	if limit > 0 {
		args = append(args, "-n", strconv.Itoa(limit))
	}
	
	if oneline {
		args = append(args, "--oneline")
	} else {
		args = append(args, "--pretty=format:%H%x1f%an%x1f%ad%x1f%s", "--date=short")
	}

	// Run git log
	logCmd := exec.Command("git", args...)
	logCmd.Dir = appPath
	var outBuf, errBuf bytes.Buffer
	logCmd.Stdout = &outBuf
	logCmd.Stderr = &errBuf
	err = logCmd.Run()
	if err != nil {
		// Check if it's just an empty repository
		errorStr := errBuf.String()
		if strings.Contains(errorStr, "does not have any commits") || strings.Contains(errorStr, "No commits yet") || strings.Contains(err.Error(), "does not have any commits") {
			return GitLogResult{
				IsRepo:      true,
				Commits:     []GitLogEntry{},
				Message:     "No commits yet",
				ErrorOutput: errorStr,
			}, nil
		}
		// Also check for exit status 128 which can indicate empty repo
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 && outBuf.Len() == 0 {
			return GitLogResult{
				IsRepo:      true,
				Commits:     []GitLogEntry{},
				Message:     "No commits yet",
				ErrorOutput: errorStr,
			}, nil
		}
		return GitLogResult{}, fmt.Errorf("failed to run git log: %w\nError output: %s", err, errorStr)
	}

	result := GitLogResult{
		IsRepo:  true,
		Commits: []GitLogEntry{},
	}

	// Parse output
	lines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		if oneline {
			// Parse oneline format
			parts := strings.SplitN(line, " ", 2)
			if len(parts) >= 2 {
				result.Commits = append(result.Commits, GitLogEntry{
					Hash:    parts[0],
					Message: parts[1],
				})
			}
		} else {
			// Parse full format
			parts := strings.Split(line, "\x1f")
			if len(parts) >= 4 {
				result.Commits = append(result.Commits, GitLogEntry{
					Hash:    parts[0][:7], // Short hash
					Author:  parts[1],
					Date:    parts[2],
					Message: parts[3],
				})
			}
		}
	}

	return result, nil
}

// CLI
func GitLogCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_log requires at least 1 argument: app_name\nUsage: layered-code tool git_log <app_name> [-n <limit>] [--oneline]")
	}

	appName := args[0]
	limit := 10 // Default limit
	oneline := false

	// Parse arguments
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-n", "--limit":
			if i+1 < len(args) {
				l, err := strconv.Atoi(args[i+1])
				if err != nil {
					return fmt.Errorf("invalid limit: %s", args[i+1])
				}
				limit = l
				i++ // Skip the limit value
			}
		case "--oneline":
			oneline = true
		}
	}

	result, err := GitLog(appName, limit, oneline)
	if err != nil {
		return fmt.Errorf("failed to get git log: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	if len(result.Commits) == 0 {
		fmt.Println("No commits yet")
		return nil
	}

	// Display commits
	for _, commit := range result.Commits {
		if oneline {
			fmt.Printf("%s %s\n", commit.Hash, commit.Message)
		} else {
			fmt.Printf("commit %s\n", commit.Hash)
			fmt.Printf("Author: %s\n", commit.Author)
			fmt.Printf("Date:   %s\n", commit.Date)
			fmt.Printf("\n    %s\n\n", commit.Message)
		}
	}

	return nil
}

// MCP
func GitLogMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
		Limit   int    `json:"limit"`
		Oneline bool   `json:"oneline"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	// Default limit if not specified
	if args.Limit == 0 {
		args.Limit = 10
	}

	result, err := GitLog(args.AppName, args.Limit, args.Oneline)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}