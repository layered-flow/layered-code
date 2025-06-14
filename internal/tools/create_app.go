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
	AppName string `json:"app_name"`
	Template string `json:"template"`
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

	// Create the .layered-code directory
	layeredDir := filepath.Join(appPath, ".layered-code")
	if err := os.MkdirAll(layeredDir, constants.AppsDirectoryPerms); err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create .layered-code directory: %v", err)}, err
	}

	// Determine template to use
	template := params.Template
	if template == "" {
		template = "vite-html" // Default to HTML template
	}

	// Validate template
	if template != "vite-html" && template != "vite-react" {
		return CreateAppResult{
			Success: false, 
			Message: fmt.Sprintf("Invalid template '%s'. Valid options are: 'vite-html' (plain HTML/JS) or 'vite-react' (React app)", template),
		}, fmt.Errorf("invalid template")
	}

	// Create a basic .layered.json config file
	layeredConfig := map[string]interface{}{
		"version":      "1.0",
		"app_name":     params.AppName,
		"project_type": "vite",
		"template":     template,
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

	// Generate port and check for conflicts
	port := GenerateUniquePort(params.AppName)
	conflictingApp, err := CheckPortConflicts(params.AppName, port)
	if err == nil && conflictingApp != "" {
		// Try a few alternative ports
		for i := 1; i <= 10; i++ {
			port = GenerateUniquePort(params.AppName + fmt.Sprintf("_%d", i))
			conflictingApp, err = CheckPortConflicts(params.AppName, port)
			if err != nil || conflictingApp == "" {
				break
			}
		}
		if conflictingApp != "" {
			return CreateAppResult{
				Success: false,
				Message: fmt.Sprintf("Could not find an available port. Port %d is already used by app '%s'", port, conflictingApp),
			}, fmt.Errorf("port conflict")
		}
	}

	// Load templates for the project type
	templateData := TemplateData{
		AppName:     params.AppName,
		AppNameSlug: CreateAppNameSlug(params.AppName),
		Port:        port,
	}

	templateFiles, err := LoadProjectTemplates(template, templateData)
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

	// Save initial runtime info with port
	runtimeInfo := RuntimeInfo{
		Port:   templateData.Port,
		Status: "created",
	}
	if err := SaveRuntimeInfo(params.AppName, runtimeInfo); err != nil {
		// Log warning but don't fail the creation
		fmt.Fprintf(os.Stderr, "Warning: failed to save runtime info: %v\n", err)
	}

	return CreateAppResult{
		Success: true,
		Message: fmt.Sprintf("Successfully created app '%s' (configured for port %d)", params.AppName, templateData.Port),
		Path:    appPath,
	}, nil
}

// CLI
func CreateAppCli() error {
	args := os.Args[3:]

	// Check for the correct number of arguments
	if len(args) < 1 || len(args) > 2 {
		fmt.Println("Error: Invalid number of arguments")
		fmt.Println()
		fmt.Println("Usage: layered-code tool create_app <app_name> [template]")
		fmt.Println()
		fmt.Println("Templates:")
		fmt.Println("  vite-html   - Plain HTML/JavaScript app (default)")
		fmt.Println("  vite-react  - React application")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  layered-code tool create_app myapp")
		fmt.Println("  layered-code tool create_app myapp vite-react")
		return fmt.Errorf("invalid arguments")
	}

	appName := args[0]
	template := "vite-html"
	if len(args) == 2 {
		template = args[1]
	}

	result, err := CreateApp(CreateAppParams{AppName: appName, Template: template})
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	fmt.Printf("%s\n", result.Message)
	fmt.Printf("Location: %s\n", result.Path)

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", result.Path)
	fmt.Printf("  npm install\n")
	fmt.Printf("  npm run dev\n")

	return nil
}

// MCP
func CreateAppMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
		Template string `json:"template"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	params := CreateAppParams{AppName: args.AppName, Template: args.Template}
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
