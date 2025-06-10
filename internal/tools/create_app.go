package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"

	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type CreateAppParams struct {
	AppName     string `json:"app_name"`
	ProjectType string `json:"project_type"`
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

	// Default to HTML if project type not specified
	if params.ProjectType == "" {
		params.ProjectType = "html"
	}

	// Validate project type
	availableTypes, err := GetAvailableProjectTypes()
	if err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to get available project types: %v", err)}, err
	}

	validType := false
	for _, t := range availableTypes {
		if t == params.ProjectType {
			validType = true
			break
		}
	}

	if !validType {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("project type must be one of: %s", strings.Join(availableTypes, ", "))}, fmt.Errorf("invalid project type: %s", params.ProjectType)
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

	// Create the standard directory structure
	directories := []string{
		filepath.Join(appPath, "src"),
		filepath.Join(appPath, "build"),
		filepath.Join(appPath, ".layered-code"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, constants.AppsDirectoryPerms); err != nil {
			return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create directory %s: %v", dir, err)}, err
		}
	}

	// Create a basic .layered.json config file
	layeredConfig := map[string]interface{}{
		"version":      "1.0",
		"app_name":     params.AppName,
		"project_type": params.ProjectType,
		"created_at":   time.Now().Format(time.RFC3339),
	}

	layeredConfigJSON, err := json.MarshalIndent(layeredConfig, "", "  ")
	if err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to marshal layered config: %v", err)}, err
	}

	layeredConfigPath := filepath.Join(appPath, ".layered.json")
	if err := os.WriteFile(layeredConfigPath, layeredConfigJSON, 0644); err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create .layered.json: %v", err)}, err
	}

	// Load templates for the project type
	templateData := TemplateData{
		AppName:     params.AppName,
		AppNameSlug: CreateAppNameSlug(params.AppName),
	}

	templateFiles, err := LoadProjectTemplates(params.ProjectType, templateData)
	if err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to load templates: %v", err)}, err
	}

	// Create files from templates
	for _, tmplFile := range templateFiles {
		filePath := filepath.Join(appPath, tmplFile.Path)

		// Ensure the directory exists
		fileDir := filepath.Dir(filePath)
		if err := os.MkdirAll(fileDir, constants.AppsDirectoryPerms); err != nil {
			return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create directory for file %s: %v", tmplFile.Path, err)}, err
		}

		// Write the file
		if err := os.WriteFile(filePath, tmplFile.Content, 0644); err != nil {
			return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create file %s: %v", tmplFile.Path, err)}, err
		}
	}

	return CreateAppResult{
		Success: true,
		Message: fmt.Sprintf("Successfully created %s app '%s' with src/build structure", params.ProjectType, params.AppName),
		Path:    appPath,
	}, nil
}

// CLI
func CreateAppCli() error {
	args := os.Args[3:]

	// Get available project types for help message
	availableTypes, err := GetAvailableProjectTypes()
	if err != nil {
		return fmt.Errorf("failed to get available project types: %w", err)
	}

	// Check for the correct number of arguments
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("create_app requires 1-2 arguments\nUsage: layered-code tool create_app <app_name> [%s]\nDefaults to 'html' if not specified", strings.Join(availableTypes, "|"))
	}

	appName := args[0]
	projectType := "html" // default

	if len(args) == 2 {
		projectType = args[1]
	}

	result, err := CreateApp(CreateAppParams{AppName: appName, ProjectType: projectType})
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	fmt.Printf("%s\n", result.Message)
	fmt.Printf("Location: %s\n", result.Path)

	if projectType == "vite" {
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  cd %s\n", result.Path)
		fmt.Printf("  npm install\n")
		fmt.Printf("  npm run dev\n")
	}

	return nil
}

// MCP
func CreateAppMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName     string `json:"app_name"`
		ProjectType string `json:"project_type"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	params := CreateAppParams{AppName: args.AppName, ProjectType: args.ProjectType}
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
