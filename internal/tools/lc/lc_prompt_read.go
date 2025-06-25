package lc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/layered-flow/layered-code/internal/tools/lc/shared"
)

// PromptRead returns the full content of a specific prompt by ID
// If no version is specified, returns the highest version
func PromptRead(identifier string) (shared.PromptReadResult, error) {
	// Check if identifier contains version (format: "id:version")
	parts := strings.Split(identifier, ":")
	requestedVersion := -1 // -1 means latest version
	
	if len(parts) == 2 {
		var v int
		if _, err := fmt.Sscanf(parts[1], "%d", &v); err == nil {
			requestedVersion = v
			identifier = parts[0]
		}
	}
	
	// Parse as numeric ID
	var promptID int
	if _, err := fmt.Sscanf(identifier, "%d", &promptID); err != nil {
		return shared.PromptReadResult{}, fmt.Errorf("invalid prompt ID '%s': must be numeric", identifier)
	}
	
	// Find the prompt with matching ID and version
	var foundPrompt *shared.Prompt
	highestVersion := 0
	
	for key, prompt := range shared.PromptsMap {
		if key.ID == promptID {
			if requestedVersion == -1 {
				// Get highest version
				if key.Version > highestVersion {
					highestVersion = key.Version
					foundPrompt = &prompt
				}
			} else if key.Version == requestedVersion {
				// Get specific version
				foundPrompt = &prompt
				break
			}
		}
	}
	
	if foundPrompt != nil {
		return shared.PromptReadResult{
			Name:        foundPrompt.Name,
			Description: foundPrompt.Description,
			Content:     foundPrompt.Content,
		}, nil
	}
	
	if requestedVersion != -1 {
		return shared.PromptReadResult{}, fmt.Errorf("prompt ID %d version %d not found", promptID, requestedVersion)
	}
	return shared.PromptReadResult{}, fmt.Errorf("prompt ID %d not found", promptID)
}

// PromptReadCli handles the CLI command for reading a specific prompt
func PromptReadCli() error {
	args := os.Args[3:]
	
	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: layered-code tool lc_prompt_read <prompt_id>[:<version>]")
			fmt.Println()
			fmt.Println("Read the content of a specific prompt by numeric ID")
			fmt.Println()
			fmt.Println("Arguments:")
			fmt.Println("  prompt_id    The numeric ID of the prompt")
			fmt.Println("  version      Optional version number (defaults to latest)")
			fmt.Println()
			fmt.Println("Version Syntax:")
			fmt.Println("  - Use 'id:version' to read a specific version")
			fmt.Println("  - Omit version to get the latest version")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  # Read latest version by ID")
			fmt.Println("  layered-code tool lc_prompt_read 1")
			fmt.Println()
			fmt.Println("  # Read specific version by ID")
			fmt.Println("  layered-code tool lc_prompt_read 1:1")
			return nil
		}
	}
	
	if len(args) < 1 {
		return fmt.Errorf("usage: layered-code tool lc_prompt_read <prompt_id>[:<version>]")
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
		PromptID string `json:"prompt_id"`
	}
	
	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}
	
	if args.PromptID == "" {
		return nil, fmt.Errorf("prompt_id is required")
	}
	
	result, err := PromptRead(args.PromptID)
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