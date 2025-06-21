package lc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/layered-flow/layered-code/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type LcListAppsResult struct {
	Apps      []string `json:"apps"`
	Directory string   `json:"directory"`
}

// LcListApps lists all applications (folders)
func LcListApps() (LcListAppsResult, error) {
	// Ensure the apps directory exists and get its path
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcListAppsResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return LcListAppsResult{}, fmt.Errorf("failed to read apps directory: %w", err)
	}

	// Filter directories only
	var apps []string
	for _, entry := range entries {
		if entry.IsDir() {
			apps = append(apps, entry.Name())
		}
	}

	// Sort apps alphabetically
	sort.Strings(apps)

	return LcListAppsResult{Apps: apps, Directory: appsDir}, nil
}

// CLI
func LcListAppsCli() error {
	args := os.Args[3:]

	// Check for any arguments (list_apps doesn't take any)
	if len(args) > 0 {
		return fmt.Errorf("lc_list_apps does not accept any arguments, got: %v", args)
	}

	result, err := LcListApps()
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	if len(result.Apps) == 0 {
		fmt.Printf("No apps found in: %s\n", result.Directory)
		return nil
	}

	fmt.Printf("Apps in '%s':\n", result.Directory)
	for _, app := range result.Apps {
		fmt.Printf("  %s\n", app)
	}
	return nil
}

// MCP
func LcListAppsMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := LcListApps()
	if err != nil {
		return nil, err
	}

	// Convert result to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
