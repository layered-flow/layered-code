package lc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// PromptRead returns the full content of a specific prompt by ID or name
func PromptRead(identifier string) (PromptReadResult, error) {
	// First try to find by ID (exact match)
	if prompt, exists := promptsMap[identifier]; exists {
		return PromptReadResult{
			Name:        prompt.Name,
			Description: prompt.Description,
			Content:     prompt.Content,
		}, nil
	}
	
	// Then try to find by name (case-insensitive)
	normalizedName := strings.ToLower(strings.ReplaceAll(identifier, " ", "_"))
	if prompt, exists := promptsMap[normalizedName]; exists {
		return PromptReadResult{
			Name:        prompt.Name,
			Description: prompt.Description,
			Content:     prompt.Content,
		}, nil
	}
	
	// Finally, search through all prompts for a case-insensitive name match
	lowerIdentifier := strings.ToLower(identifier)
	for _, prompt := range promptsMap {
		if strings.ToLower(prompt.Name) == lowerIdentifier {
			return PromptReadResult{
				Name:        prompt.Name,
				Description: prompt.Description,
				Content:     prompt.Content,
			}, nil
		}
	}
	
	return PromptReadResult{}, fmt.Errorf("prompt '%s' not found", identifier)
}

// PromptReadCli handles the CLI command for reading a specific prompt
func PromptReadCli() error {
	args := os.Args[3:]
	
	if len(args) < 1 {
		return fmt.Errorf("usage: layered-code tool lc_prompt_read <prompt_id_or_name>")
	}
	
	promptName := strings.Join(args, " ")
	
	result, err := PromptRead(promptName)
	if err != nil {
		return fmt.Errorf("failed to read prompt: %w", err)
	}
	
	fmt.Printf("# %s\n\n", result.Name)
	fmt.Printf("Description: %s\n\n", result.Description)
	fmt.Println("---")
	fmt.Println()
	fmt.Println(result.Content)
	
	return nil
}

// PromptReadMcp handles the MCP request for reading a specific prompt
func PromptReadMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		PromptName string `json:"prompt_name"`
	}
	
	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}
	
	if args.PromptName == "" {
		return nil, fmt.Errorf("prompt_name is required")
	}
	
	result, err := PromptRead(args.PromptName)
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