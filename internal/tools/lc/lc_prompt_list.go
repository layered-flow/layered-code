package lc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

// PromptList returns a list of all available prompts with their names and descriptions
// Only the highest version of each prompt ID is shown
func PromptList() (PromptListResult, error) {
	// Create a map to track the highest version for each prompt ID
	highestVersions := make(map[int]Prompt)
	
	for _, prompt := range promptsMap {
		existing, exists := highestVersions[prompt.ID]
		if !exists || prompt.Version > existing.Version {
			highestVersions[prompt.ID] = prompt
		}
	}
	
	// Convert map to slice
	prompts := make([]Prompt, 0, len(highestVersions))
	for _, prompt := range highestVersions {
		// Create a copy without content for listing
		prompts = append(prompts, Prompt{
			ID:          prompt.ID,
			Version:     prompt.Version,
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
	// Check for help flag
	if len(os.Args) > 3 {
		for _, arg := range os.Args[3:] {
			if arg == "--help" || arg == "-h" {
				fmt.Println("Usage: layered-code tool lc_prompt_list")
				fmt.Println()
				fmt.Println("List all available prompts with their IDs, versions, and descriptions")
				fmt.Println()
				fmt.Println("Notes:")
				fmt.Println("  - Only the highest version of each prompt is shown")
				fmt.Println("  - Use lc_prompt_read to view the full content of a prompt")
				fmt.Println()
				fmt.Println("Examples:")
				fmt.Println("  # List all prompts")
				fmt.Println("  layered-code tool lc_prompt_list")
				return nil
			}
		}
	}
	
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
		fmt.Printf("• %s (ID: %d, Version: %d)\n", prompt.Name, prompt.ID, prompt.Version)
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