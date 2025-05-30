package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/tools/git"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// Types
type LcmListEntry struct {
	Index         int    `json:"index"`
	Timestamp     string `json:"timestamp"`
	Commit        string `json:"commit"`
	Summary       string `json:"summary"`
}

type LcmListResult struct {
	Success bool           `json:"success"`
	Entries []LcmListEntry `json:"entries"`
	Total   int            `json:"total"`
	Message string         `json:"message,omitempty"`
}

// LcmList reads the LayeredChangeMemory YAML file and returns summaries
func LcmList(appName string) (LcmListResult, error) {
	if err := git.ValidateAppName(appName); err != nil {
		return LcmListResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcmListResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := git.ValidateAppPath(appPath); err != nil {
		return LcmListResult{}, err
	}

	// Check if LCM file exists
	lcmPath := filepath.Join(appPath, ".layered_change_memory.yaml")
	if _, err := os.Stat(lcmPath); os.IsNotExist(err) {
		return LcmListResult{
			Success: true,
			Entries: []LcmListEntry{},
			Total:   0,
			Message: "No layered change memory found",
		}, nil
	}

	// Read the YAML file
	data, err := os.ReadFile(lcmPath)
	if err != nil {
		return LcmListResult{}, fmt.Errorf("failed to read LCM file: %w", err)
	}

	// Parse YAML
	var entries []git.LayeredChangeMemoryEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return LcmListResult{}, fmt.Errorf("failed to parse LCM file: %w", err)
	}

	// Convert to list entries with index
	listEntries := make([]LcmListEntry, len(entries))
	for i, entry := range entries {
		listEntries[i] = LcmListEntry{
			Index:     i,
			Timestamp: entry.Timestamp,
			Commit:    entry.Commit,
			Summary:   entry.Summary,
		}
	}

	return LcmListResult{
		Success: true,
		Entries: listEntries,
		Total:   len(listEntries),
	}, nil
}

// CLI
func LcmListCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("lcm_list requires 1 argument: app_name\nUsage: layered-code tool lcm_list <app_name>")
	}

	appName := args[0]

	result, err := LcmList(appName)
	if err != nil {
		return fmt.Errorf("failed to list LCM entries: %w", err)
	}

	if result.Message != "" {
		fmt.Println(result.Message)
		return nil
	}

	if len(result.Entries) == 0 {
		fmt.Println("No layered change memory entries found")
		return nil
	}

	fmt.Printf("Found %d layered change memory entries:\n\n", result.Total)
	for _, entry := range result.Entries {
		fmt.Printf("[%d] %s | %s\n    %s\n\n", entry.Index, entry.Timestamp, entry.Commit, entry.Summary)
	}

	return nil
}

// MCP
func LcmListMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := LcmList(args.AppName)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}