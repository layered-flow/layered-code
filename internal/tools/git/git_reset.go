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

// ResetMode represents the git reset mode
type ResetMode string

const (
	ResetModeSoft  ResetMode = "soft"
	ResetModeMixed ResetMode = "mixed"
	ResetModeHard  ResetMode = "hard"
)

// Reset performs a git reset operation
func Reset(appName, commitHash string, mode ResetMode) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("app_name is required")
	}
	if commitHash == "" {
		return "", fmt.Errorf("commit_hash is required")
	}

	// Validate mode
	if mode != "" && mode != ResetModeSoft && mode != ResetModeMixed && mode != ResetModeHard {
		return "", fmt.Errorf("invalid reset mode: %s (must be 'soft', 'mixed', or 'hard')", mode)
	}

	// Default to mixed if not specified
	if mode == "" {
		mode = ResetModeMixed
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return "", fmt.Errorf("failed to get apps directory: %w", err)
	}
	repoPath := filepath.Join(appsDir, appName)
	
	// Build command
	args := []string{"reset", fmt.Sprintf("--%s", mode), commitHash}
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git reset failed: %w\nOutput: %s", err, string(output))
	}

	// Get the current status after reset
	statusCmd := exec.Command("git", "status", "--porcelain", "-uno")
	statusCmd.Dir = repoPath
	statusOutput, _ := statusCmd.CombinedOutput()

	result := fmt.Sprintf("Successfully reset to commit %s using %s mode", commitHash, mode)
	
	if len(statusOutput) > 0 {
		result += fmt.Sprintf("\n\nModified files:\n%s", string(statusOutput))
	}

	// Also show the new HEAD commit
	logCmd := exec.Command("git", "log", "--oneline", "-1")
	logCmd.Dir = repoPath
	if logOutput, err := logCmd.CombinedOutput(); err == nil {
		result += fmt.Sprintf("\n\nNow at: %s", strings.TrimSpace(string(logOutput)))
	}

	return result, nil
}

// CLI
func GitResetCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("git_reset requires at least 2 arguments: app_name commit_hash [mode]\nUsage: layered-code tool git_reset <app_name> <commit_hash> [soft|mixed|hard]")
	}

	appName := args[0]
	commitHash := args[1]
	
	mode := ResetModeMixed // default
	if len(args) > 2 {
		switch args[2] {
		case "soft":
			mode = ResetModeSoft
		case "mixed":
			mode = ResetModeMixed
		case "hard":
			mode = ResetModeHard
		default:
			return fmt.Errorf("invalid reset mode: %s (must be 'soft', 'mixed', or 'hard')", args[2])
		}
	}

	result, err := Reset(appName, commitHash, mode)
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	fmt.Println(result)
	return nil
}

// MCP
func GitResetMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName    string `json:"app_name"`
		CommitHash string `json:"commit_hash"`
		Mode       string `json:"mode,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Convert string mode to ResetMode
	mode := ResetModeMixed
	if args.Mode != "" {
		switch args.Mode {
		case "soft":
			mode = ResetModeSoft
		case "mixed":
			mode = ResetModeMixed
		case "hard":
			mode = ResetModeHard
		default:
			return nil, fmt.Errorf("invalid reset mode: %s", args.Mode)
		}
	}

	result, err := Reset(args.AppName, args.CommitHash, mode)
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