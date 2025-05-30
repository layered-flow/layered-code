package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/tools/git"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// Types
type LcmReadResult struct {
	Success bool                          `json:"success"`
	Entry   *git.LayeredChangeMemoryEntry `json:"entry,omitempty"`
	Message string                        `json:"message,omitempty"`
}

// LcmRead reads a specific LayeredChangeMemory entry from the YAML file
func LcmRead(appName string, index int) (LcmReadResult, error) {
	if err := git.ValidateAppName(appName); err != nil {
		return LcmReadResult{}, err
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcmReadResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	appPath := filepath.Join(appsDir, appName)
	if err := git.ValidateAppPath(appPath); err != nil {
		return LcmReadResult{}, err
	}

	// Check if LCM file exists
	lcmPath := filepath.Join(appPath, ".layered_change_memory.yaml")
	if _, err := os.Stat(lcmPath); os.IsNotExist(err) {
		return LcmReadResult{
			Success: false,
			Message: "No layered change memory found",
		}, nil
	}

	// Read the YAML file
	data, err := os.ReadFile(lcmPath)
	if err != nil {
		return LcmReadResult{}, fmt.Errorf("failed to read LCM file: %w", err)
	}

	// Parse YAML
	var entries []git.LayeredChangeMemoryEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return LcmReadResult{}, fmt.Errorf("failed to parse LCM file: %w", err)
	}

	// Validate index
	if index < 0 || index >= len(entries) {
		return LcmReadResult{
			Success: false,
			Message: fmt.Sprintf("Invalid index %d. Valid range is 0-%d", index, len(entries)-1),
		}, nil
	}

	// Return the requested entry
	return LcmReadResult{
		Success: true,
		Entry:   &entries[index],
	}, nil
}

// CLI
func LcmReadCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("lcm_read requires 2 arguments: app_name index\nUsage: layered-code tool lcm_read <app_name> <index>")
	}

	appName := args[0]
	indexStr := args[1]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return fmt.Errorf("invalid index: %s (must be a number)", indexStr)
	}

	result, err := LcmRead(appName, index)
	if err != nil {
		return fmt.Errorf("failed to read LCM entry: %w", err)
	}

	if !result.Success {
		fmt.Println(result.Message)
		return nil
	}

	if result.Entry == nil {
		fmt.Println("Entry not found")
		return nil
	}

	// Display the entry in a readable format
	entry := result.Entry
	fmt.Printf("Layered Change Memory Entry [%d]:\n", index)
	fmt.Printf("=====================================\n")
	fmt.Printf("Timestamp:      %s\n", entry.Timestamp)
	fmt.Printf("Commit:         %s\n", entry.Commit)
	fmt.Printf("Commit Message: %s\n", entry.CommitMessage)
	fmt.Printf("Summary:        %s\n", entry.Summary)
	
	if len(entry.Considerations) > 0 {
		fmt.Printf("Considerations:\n")
		for i, consideration := range entry.Considerations {
			fmt.Printf("  %d. %s\n", i+1, consideration)
		}
	}
	
	if entry.FollowUp != "" {
		fmt.Printf("Follow-up:      %s\n", entry.FollowUp)
	}

	return nil
}

// MCP
func LcmReadMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
		Index   int    `json:"index"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if args.AppName == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	result, err := LcmRead(args.AppName, args.Index)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}