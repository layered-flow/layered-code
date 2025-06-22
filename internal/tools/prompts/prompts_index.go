package prompts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/layered-flow/layered-code/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type Prompt struct {
	Name        string `json:"name"`
	Body 				string `json:"body"`
}

type PromptsIndexResult struct {
	Prompts   []Prompt `json:"prompts"`
	Directory string   `json:"directory"`
}

// PromptsIndex lists all available prompts
func PromptsIndex() (PromptsIndexResult, error) {
	// Get the prompts directory path
	promptsDir, err := getPromptsDirectory()
	if err != nil {
		return PromptsIndexResult{}, fmt.Errorf("failed to get prompts directory: %w", err)
	}

	// For now, return sample prompts
	// Later these can be loaded from files in the prompts directory
	prompts := []Prompt{
		{
			Name: "general_requirements",
			Body:	"You are a collaborative assistant for building applications. Begin by asking the user what they want to achieve. Then work with them step by step to help reach that goal. Start by creating a plan and presenting it for the user’s approval. Only proceed after they confirm the plan. For each step, prefer small, incremental changes so the user has regular opportunities to review and give feedback. Stop if any unexpected issues arise, and clearly explain the problem. Offer the user several options for how to proceed. Only commit changes when the user explicitly approves them and confirms they’re satisfied with the result. Always default to careful, user-guided progress unless instructed otherwise.",
		},
		{
			Name: "create_react_app",
			Body:	"Gather requirements: app name, UI library (Radix UI, Tailwind, Bootstrap, Material-UI). Dark theme, mobile responsive.",
		},
	}

	// Sort prompts alphabetically by name
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})

	return PromptsIndexResult{Prompts: prompts, Directory: promptsDir}, nil
}

// getPromptsDirectory returns the prompts directory path within the apps directory
func getPromptsDirectory() (string, error) {
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return "", err
	}

	// Create a dedicated prompts directory within the apps directory
	promptsDir := filepath.Join(appsDir, ".prompts")
	
	// Ensure the directory exists
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create prompts directory: %w", err)
	}

	return promptsDir, nil
}

// CLI
func PromptsIndexCli() error {
	args := os.Args[3:]

	// Check for any arguments (prompts_index doesn't take any)
	if len(args) > 0 {
		return fmt.Errorf("prompts_index does not accept any arguments, got: %v", args)
	}

	result, err := PromptsIndex()
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	if len(result.Prompts) == 0 {
		fmt.Printf("No prompts found in: %s\n", result.Directory)
		return nil
	}

	fmt.Printf("Available prompts in '%s':\n", result.Directory)
	for _, prompt := range result.Prompts {
		fmt.Printf("  %-20s - %s\n", prompt.Name, prompt.Body)				
	}
	return nil
}

// MCP
func PromptsIndexMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := PromptsIndex()
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