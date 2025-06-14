package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// AppInfoParams represents the parameters for getting app info
type AppInfoParams struct {
	AppName string `json:"app_name"`
}

// AppInfoResult represents detailed information about an app
type AppInfoResult struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
	AppName     string      `json:"app_name,omitempty"`
	Path        string      `json:"path,omitempty"`
	Port        int         `json:"port,omitempty"`
	Status      string      `json:"status,omitempty"`
	ProjectType string      `json:"project_type,omitempty"`
	Template    string      `json:"template,omitempty"`
	CreatedAt   string      `json:"created_at,omitempty"`
	HasNodeModules bool     `json:"has_node_modules,omitempty"`
	URL         string      `json:"url,omitempty"`
}

// GetAppInfo retrieves detailed information about an app
func GetAppInfo(params AppInfoParams) (AppInfoResult, error) {
	// Validate app name
	if err := validateAppName(params.AppName); err != nil {
		return AppInfoResult{Success: false, Message: err.Error()}, err
	}

	// Get the app directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return AppInfoResult{Success: false, Message: fmt.Sprintf("failed to get apps directory: %v", err)}, err
	}

	appPath := filepath.Join(appsDir, params.AppName)

	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return AppInfoResult{Success: false, Message: fmt.Sprintf("app '%s' does not exist", params.AppName)}, fmt.Errorf("app not found")
	}

	result := AppInfoResult{
		Success: true,
		AppName: params.AppName,
		Path:    appPath,
	}

	// Read .layered.json
	layeredPath := filepath.Join(appPath, ".layered.json")
	if data, err := os.ReadFile(layeredPath); err == nil {
		var layeredConfig map[string]interface{}
		if json.Unmarshal(data, &layeredConfig) == nil {
			if pt, ok := layeredConfig["project_type"].(string); ok {
				result.ProjectType = pt
			}
			if tmpl, ok := layeredConfig["template"].(string); ok {
				result.Template = tmpl
			}
			if ca, ok := layeredConfig["created_at"].(string); ok {
				result.CreatedAt = ca
			}
		}
	}

	// Get port and runtime information
	port := GetAppPort(appPath)
	result.Port = port

	// Get runtime info for status
	if runtimeInfo, err := GetRuntimeInfo(params.AppName); err == nil {
		result.Status = runtimeInfo.Status
		if runtimeInfo.Port > 0 {
			result.Port = runtimeInfo.Port
		}
	}

	// Check if node_modules exists
	if _, err := os.Stat(filepath.Join(appPath, "node_modules")); err == nil {
		result.HasNodeModules = true
	}

	// Build URL if app is running
	if result.Status == "running" && result.Port > 0 {
		result.URL = fmt.Sprintf("http://localhost:%d", result.Port)
	}

	result.Message = fmt.Sprintf("App '%s' information retrieved successfully", params.AppName)
	return result, nil
}

// CLI
func AppInfoCli() error {
	args := os.Args[3:]

	if len(args) != 1 {
		return fmt.Errorf("app_info requires 1 argument\nUsage: layered-code tool app_info <app_name>")
	}

	appName := args[0]
	result, err := GetAppInfo(AppInfoParams{AppName: appName})
	if err != nil {
		return fmt.Errorf("failed to get app info: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	// Print app information
	fmt.Printf("App: %s\n", result.AppName)
	fmt.Printf("Path: %s\n", result.Path)
	fmt.Printf("Port: %d\n", result.Port)
	if result.Status != "" {
		fmt.Printf("Status: %s\n", result.Status)
	}
	if result.ProjectType != "" {
		fmt.Printf("Project Type: %s\n", result.ProjectType)
	}
	if result.Template != "" {
		fmt.Printf("Template: %s\n", result.Template)
	}
	if result.URL != "" {
		fmt.Printf("URL: %s\n", result.URL)
	}
	fmt.Printf("Node Modules: %v\n", result.HasNodeModules)

	return nil
}

// MCP
func AppInfoMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	params := AppInfoParams{AppName: args.AppName}
	result, err := GetAppInfo(params)
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