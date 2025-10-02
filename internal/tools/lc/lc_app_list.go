package lc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type AppInfo struct {
	Name        string `json:"name"`
	HasAgentsMd bool   `json:"has_agents_md"`
}

type LcAppListResult struct {
	Apps      []AppInfo `json:"apps"`
	Directory string    `json:"directory"`
}

// LcAppList lists all applications (folders) and checks for AGENTS.md
func LcAppList() (LcAppListResult, error) {
	// Ensure the apps directory exists and get its path
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcAppListResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return LcAppListResult{}, fmt.Errorf("failed to read apps directory: %w", err)
	}

	// Filter directories only and check for AGENTS.md
	var apps []AppInfo
	for _, entry := range entries {
		if entry.IsDir() {
			appInfo := AppInfo{
				Name:        entry.Name(),
				HasAgentsMd: hasAgentsMdFile(filepath.Join(appsDir, entry.Name())),
			}
			apps = append(apps, appInfo)
		}
	}

	// Sort apps alphabetically by name
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Name < apps[j].Name
	})

	return LcAppListResult{Apps: apps, Directory: appsDir}, nil
}

// hasAgentsMdFile checks if an AGENTS.md file exists in the directory (case-insensitive)
func hasAgentsMdFile(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.ToLower(entry.Name()) == "agents.md" {
			return true
		}
	}
	return false
}

// CLI
func LcAppListCli() error {
	args := os.Args[3:]

	// Check for any arguments (list_apps doesn't take any)
	if len(args) > 0 {
		return fmt.Errorf("lc_app_list does not accept any arguments, got: %v", args)
	}

	result, err := LcAppList()
	if err != nil {
		return fmt.Errorf("failed to list apps: %w", err)
	}

	if len(result.Apps) == 0 {
		fmt.Printf("No apps found in: %s\n", result.Directory)
		return nil
	}

	fmt.Printf("Apps in '%s':\n", result.Directory)
	for _, app := range result.Apps {
		agentsMdIndicator := ""
		if app.HasAgentsMd {
			agentsMdIndicator = " [AGENTS.md]"
		}
		fmt.Printf("  %s%s\n", app.Name, agentsMdIndicator)
	}
	return nil
}

// MCP
func LcAppListMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := LcAppList()
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
