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
type GitRemoteResult struct {
	Remotes       map[string]string `json:"remotes"`
	IsRepo        bool              `json:"is_repo"`
	Message       string            `json:"message,omitempty"`
	AddSuccess    bool              `json:"add_success,omitempty"`
	RemoveSuccess bool              `json:"remove_success,omitempty"`
	RenameSuccess bool              `json:"rename_success,omitempty"`
	ErrorOutput   string            `json:"error_output,omitempty"`
}

// GitRemote manages git remotes in the specified app directory
func GitRemote(appName string, addName string, addURL string, removeName string, oldName string, newName string, setURL string, setURLName string) (GitRemoteResult, error) {
	if err := EnsureGitAvailable(); err != nil {
		return GitRemoteResult{}, err
	}

	if err := helpers.ValidateAppName(appName); err != nil {
		return GitRemoteResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return GitRemoteResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := ValidateAppPath(appPath); err != nil {
		return GitRemoteResult{}, err
	}

	// Check if it's a git repository
	gitDir := filepath.Join(appPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return GitRemoteResult{
			IsRepo:  false,
			Message: fmt.Sprintf("'%s' is not a git repository", appName),
		}, nil
	}

	result := GitRemoteResult{
		IsRepo:  true,
		Remotes: make(map[string]string),
	}

	// Handle add remote
	if addName != "" && addURL != "" {
		cmd := exec.Command("git", "remote", "add", addName, addURL)
		cmd.Dir = appPath
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err := cmd.Run()
		if err != nil {
			result.Message = fmt.Sprintf("Failed to add remote: %s", strings.TrimSpace(errBuf.String()))
			result.AddSuccess = false
			result.ErrorOutput = errBuf.String()
		} else {
			result.AddSuccess = true
		}
	}

	// Handle remove remote
	if removeName != "" {
		cmd := exec.Command("git", "remote", "remove", removeName)
		cmd.Dir = appPath
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err := cmd.Run()
		if err != nil {
			result.Message = fmt.Sprintf("Failed to remove remote: %s", strings.TrimSpace(errBuf.String()))
			result.RemoveSuccess = false
			result.ErrorOutput = errBuf.String()
		} else {
			result.RemoveSuccess = true
		}
	}

	// Handle rename remote
	if oldName != "" && newName != "" {
		cmd := exec.Command("git", "remote", "rename", oldName, newName)
		cmd.Dir = appPath
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err := cmd.Run()
		if err != nil {
			result.Message = fmt.Sprintf("Failed to rename remote: %s", strings.TrimSpace(errBuf.String()))
			result.RenameSuccess = false
			result.ErrorOutput = errBuf.String()
		} else {
			result.RenameSuccess = true
		}
	}

	// Handle set-url
	if setURLName != "" && setURL != "" {
		cmd := exec.Command("git", "remote", "set-url", setURLName, setURL)
		cmd.Dir = appPath
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err := cmd.Run()
		if err != nil {
			result.Message = fmt.Sprintf("Failed to set remote URL: %s", strings.TrimSpace(errBuf.String()))
			result.ErrorOutput = errBuf.String()
		}
	}

	// List all remotes with their URLs
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = appPath
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	if err != nil {
		if result.ErrorOutput == "" {
			result.ErrorOutput = errBuf.String()
		}
		return result, nil
	}

	// Parse remote list
	lines := strings.Split(outBuf.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse "origin  https://github.com/user/repo.git (fetch)" format
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			remoteName := parts[0]
			remoteURL := parts[1]
			// Only add fetch URLs to avoid duplicates
			if len(parts) >= 3 && strings.Contains(parts[2], "fetch") {
				result.Remotes[remoteName] = remoteURL
			}
		}
	}

	return result, nil
}

// CLI
func GitRemoteCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("git_remote requires at least 1 argument: app_name\nUsage: layered-code tool git_remote <app_name> [-a <name> <url>] [-r <name>] [--rename <old> <new>] [--set-url <name> <url>]")
	}

	appName := args[0]
	addName := ""
	addURL := ""
	removeName := ""
	oldName := ""
	newName := ""
	setURLName := ""
	setURL := ""

	// Parse arguments
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-a", "--add":
			if i+2 < len(args) {
				addName = args[i+1]
				addURL = args[i+2]
				i += 2
			}
		case "-r", "--remove":
			if i+1 < len(args) {
				removeName = args[i+1]
				i++
			}
		case "--rename":
			if i+2 < len(args) {
				oldName = args[i+1]
				newName = args[i+2]
				i += 2
			}
		case "--set-url":
			if i+2 < len(args) {
				setURLName = args[i+1]
				setURL = args[i+2]
				i += 2
			}
		}
	}

	result, err := GitRemote(appName, addName, addURL, removeName, oldName, newName, setURL, setURLName)
	if err != nil {
		return fmt.Errorf("failed to manage remotes: %w", err)
	}

	if !result.IsRepo {
		fmt.Println(result.Message)
		return nil
	}

	// Show operation results
	if addName != "" {
		if result.AddSuccess {
			fmt.Printf("Added remote '%s' with URL: %s\n", addName, addURL)
		} else {
			fmt.Printf("Failed to add remote: %s\n", result.Message)
		}
	}

	if removeName != "" {
		if result.RemoveSuccess {
			fmt.Printf("Removed remote: %s\n", removeName)
		} else {
			fmt.Printf("Failed to remove remote: %s\n", result.Message)
		}
	}

	if oldName != "" && newName != "" {
		if result.RenameSuccess {
			fmt.Printf("Renamed remote from '%s' to '%s'\n", oldName, newName)
		} else {
			fmt.Printf("Failed to rename remote: %s\n", result.Message)
		}
	}

	if setURLName != "" && setURL != "" {
		if result.Message == "" {
			fmt.Printf("Updated URL for remote '%s' to: %s\n", setURLName, setURL)
		} else {
			fmt.Printf("Failed to set remote URL: %s\n", result.Message)
		}
	}

	// Show remote list if no operation was performed
	if addName == "" && removeName == "" && oldName == "" && setURLName == "" {
		fmt.Printf("Git Remotes for '%s':\n", appName)
		if len(result.Remotes) == 0 {
			fmt.Println("No remotes configured")
		} else {
			for name, url := range result.Remotes {
				fmt.Printf("  %s\t%s\n", name, url)
			}
		}
	}

	return nil
}

// MCP
func GitRemoteMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName     string `json:"app_name"`
		AddName     string `json:"add_name"`
		AddURL      string `json:"add_url"`
		RemoveName  string `json:"remove_name"`
		OldName     string `json:"old_name"`
		NewName     string `json:"new_name"`
		SetURLName  string `json:"set_url_name"`
		SetURL      string `json:"set_url"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := GitRemote(args.AppName, args.AddName, args.AddURL, args.RemoveName, args.OldName, args.NewName, args.SetURL, args.SetURLName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}