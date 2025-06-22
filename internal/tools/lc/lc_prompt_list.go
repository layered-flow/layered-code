package lc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// PromptList returns a list of all available prompts with their names and descriptions
func PromptList() (PromptListResult, error) {
	prompts := make([]Prompt, 0, len(promptsMap))
	
	for _, prompt := range promptsMap {
		// Create a copy without content for listing
		prompts = append(prompts, Prompt{
			ID:          prompt.ID,
			Name:        prompt.Name,
			Description: prompt.Description,
		})
	}
	
	return PromptListResult{
		Prompts: prompts,
		Count:   len(prompts),
	}, nil
}

// PromptListCli handles the CLI command for listing prompts
func PromptListCli() error {
	result, err := PromptList()
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}
	
	if result.Count == 0 {
		fmt.Println("No prompts available")
		return nil
	}
	
	fmt.Printf("Available prompts (%d):\n\n", result.Count)
	for _, prompt := range result.Prompts {
		fmt.Printf("• %s (ID: %d)\n", prompt.Name, prompt.ID)
		fmt.Printf("  %s\n\n", prompt.Description)
	}
	
	return nil
}

// PromptListMcp handles the MCP request for listing prompts
func PromptListMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := PromptList()
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