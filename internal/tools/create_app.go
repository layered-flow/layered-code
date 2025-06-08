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

	// Create .gitignore file
	gitignoreContent := `.layered-code/
build/
dist/
node_modules/
.DS_Store
*.log
.env
.env.local
`
	gitignorePath := filepath.Join(appPath, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create .gitignore: %v", err)}, err
	}

	// Create a basic .layered.json config file
	layeredConfig := map[string]interface{}{
		"version":    "1.0",
		"app_name":   params.AppName,
		"structure":  "src-build",
		"created_at": time.Now().Format(time.RFC3339),
	}

	layeredConfigJSON, err := json.MarshalIndent(layeredConfig, "", "  ")
	if err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to marshal layered config: %v", err)}, err
	}

	layeredConfigPath := filepath.Join(appPath, ".layered.json")
	if err := os.WriteFile(layeredConfigPath, layeredConfigJSON, 0644); err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create .layered.json: %v", err)}, err
	}

	// Create a simple starter index.html in src
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + params.AppName + `</title>
</head>
<body>
    <h1>` + params.AppName + `</h1>
    <p>Welcome to your new app created with Layered Code.</p>
</body>
</html>`

	indexPath := filepath.Join(appPath, "src", "index.html")
	if err := os.WriteFile(indexPath, []byte(indexHTML), 0644); err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create index.html: %v", err)}, err
	}

	// Create README.md
	readmeContent := `# ` + params.AppName + `

This app was created with Layered Code.

## Structure

- **src/** - Source code for your application
- **build/** - Compiled/built files for deployment (gitignored)
- **.layered-code/** - Layered Code metadata (gitignored)
- **.layered.json** - Layered Code configuration

## Getting Started

1. Add your source code to the ` + "`src/`" + ` directory
2. Build your app (output to ` + "`build/`" + ` directory)
3. Deploy only the contents of ` + "`build/`" + ` to your web server

## Notes

The ` + "`.layered-code/`" + ` directory and ` + "`build/`" + ` directory are gitignored by default to keep your repository clean and prevent accidental deployment of development files.
`

	readmePath := filepath.Join(appPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return CreateAppResult{Success: false, Message: fmt.Sprintf("failed to create README.md: %v", err)}, err
	}

	return CreateAppResult{
		Success: true,
		Message: fmt.Sprintf("Successfully created app '%s' with src/build structure", params.AppName),
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
