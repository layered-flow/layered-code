package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"

	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type CreateAppParams struct {
	AppName string `json:"app_name"`
}

type CreateAppResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

// validateAppName validates the app name
func validateAppName(appName string) error {
	if appName == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(appName, "/\\:*?\"<>|") {
		return fmt.Errorf("app name contains invalid characters")
	}

	// Check for directory traversal attempts
	if strings.Contains(appName, "..") {
		return fmt.Errorf("app name cannot contain '..'")
	}

	// Check for hidden directories
	if strings.HasPrefix(appName, ".") {
		return fmt.Errorf("app name cannot start with '.'")
	}

	return nil
}

// CreateApp creates a new application directory
func CreateApp(params CreateAppParams) (CreateAppResult, error) {
	// Validate app name
	if err := validateAppName(params.AppName); err != nil {
		return CreateAppResult{Success: false, Message: err.Error()}, err
	}

	// Ensure the apps directory exists and get its path
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to ensure apps directory: %v", err)}, err
	}

	// Create the full path for the new app
	appPath := filepath.Join(appsDir, params.AppName)

	// Check if the app already exists
	if _, err := os.Stat(appPath); err == nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("app '%s' already exists", params.AppName)}, fmt.Errorf("app '%s' already exists", params.AppName)
	} else if !os.IsNotExist(err) {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to check if app exists: %v", err)}, err
	}

	// Create the app directory
	if err := os.MkdirAll(appPath, constants.AppsDirectoryPerms); err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create app directory: %v", err)}, err
	}

	return CreateAppResult{
		Success: true,
		Message: fmt.Sprintf("Successfully created app '%s'", params.AppName),
		Path:    appPath,
	}, nil
}

// CLI
func CreateAppCli() error {
	args := os.Args[3:]

	// Check for the correct number of arguments
	if len(args) != 1 {
		return fmt.Errorf("create_app requires exactly one argument: the app name\nUsage: layered-code tool create_app <app_name>")
	}

	appName := args[0]

	result, err := CreateApp(CreateAppParams{AppName: appName})
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	fmt.Printf("%s\n", result.Message)
	fmt.Printf("Location: %s\n", result.Path)
	return nil
}

// MCP
func CreateAppMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	params := CreateAppParams{AppName: args.AppName}
	result, err := CreateApp(params)
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