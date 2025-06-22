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
	Description string `json:"description"`
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
			Name:        "create_html_app",
			Description: "Collaboratively creates minimal Vite HTML app. First ask user's technical level. Gather requirements: app name, TypeScript, CSS framework (Radix UI, Tailwind, Bootstrap, Material-UI). Create plan and get approval before coding. Dark theme, mobile-responsive. Only commit after confirms",
		},
		{
			Name:        "create_react_app",
			Description: "Collaboratively creates minimal React TypeScript app. First ask user's technical level. Gather requirements: app name, UI library (Radix UI, Tailwind, Bootstrap, Material-UI). Present plan for approval before coding. Dark theme, mobile responsive. Only commit after approval",
		},
		{
			Name:        "create_vue_app",
			Description: "Collaboratively creates minimal Vue 3 app. First ask user's technical level. Gather requirements: app name, UI framework (Radix UI, Tailwind, Bootstrap, Material-UI). Share plan and get feedback before coding. Dark theme, mobile-first. Only commit after verifies",
		},
		{
			Name:        "create_fullstack_app",
			Description: "Collaboratively creates minimal full-stack app. First ask user's technical level. Gather requirements: app name, frontend/backend, database. Present architecture plan for approval before coding. Dark theme UI (Radix UI, Tailwind, Bootstrap, Material-UI). Only commit after confirms",
		},
		{
			Name:        "add_component_library",
			Description: "Collaboratively adds UI library. First ask user's technical level. Gather requirements: which app, library (Radix UI, Tailwind, Bootstrap, Material-UI). Present integration plan before coding. Get approval for components. Dark theme, mobile responsive. Only commit after verifies",
		},
		{
			Name:        "setup_testing_framework",
			Description: "Collaboratively sets up minimal testing. First ask user's technical level. Gather requirements: which app, test runner (Vitest/Jest), E2E needs. Present testing strategy for approval before coding. Only essential packages. Only commit after user runs tests and approves",
		},
		{
			Name:        "deploy_app_setup",
			Description: "Collaboratively prepares deployment. First ask user's technical level. Gather requirements: which app, deployment target (Vercel/Netlify/AWS/Docker). Present deployment plan for review before creating configs. Only required configs. Only commit after user reviews settings",
		},
		{
			Name:        "search_and_refactor",
			Description: "Collaboratively refactors code. First ask user's technical level. Gather requirements: which app, search pattern, replacement. Show matches and refactoring plan for approval before making changes. Only requested changes. Only commit after user tests and confirms",
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
		fmt.Printf("  %-20s - %s\n", prompt.Name, prompt.Description)
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