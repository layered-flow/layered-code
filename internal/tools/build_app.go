package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type BuildAppParams struct {
	AppName string `json:"app_name"`
}

type BuildAppResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// BuildApp builds the application using npm/pnpm
func BuildApp(params BuildAppParams) (BuildAppResult, error) {
	startTime := time.Now()

	// Validate app name
	if err := validateAppName(params.AppName); err != nil {
		return BuildAppResult{Success: false, Message: err.Error()}, err
	}

	// Get the app directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return BuildAppResult{Success: false, Message: fmt.Sprintf("failed to get apps directory: %v", err)}, err
	}

	appPath := filepath.Join(appsDir, params.AppName)

	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return BuildAppResult{Success: false, Message: fmt.Sprintf("app '%s' does not exist", params.AppName)}, fmt.Errorf("app not found")
	}

	// Check if package.json exists
	packageJsonPath := filepath.Join(appPath, "package.json")
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		return BuildAppResult{Success: false, Message: "package.json not found"}, fmt.Errorf("package.json not found")
	}

	// Read package.json to check if build script exists
	packageData, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return BuildAppResult{Success: false, Message: fmt.Sprintf("failed to read package.json: %v", err)}, err
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(packageData, &packageJSON); err != nil {
		return BuildAppResult{Success: false, Message: fmt.Sprintf("failed to parse package.json: %v", err)}, err
	}

	// Check if build script exists
	scripts, ok := packageJSON["scripts"].(map[string]interface{})
	if !ok || scripts["build"] == nil {
		return BuildAppResult{Success: false, Message: "no 'build' script found in package.json"}, fmt.Errorf("build script not found")
	}

	// Detect package manager
	pm := detectPackageManager(appPath, "")

	// Build the command
	cmdArgs := []string{"run", "build"}
	cmd := exec.Command(string(pm), cmdArgs...)
	cmd.Dir = appPath

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment variables
	cmd.Env = append(os.Environ(), "NODE_ENV=production")

	// Run the build command
	err = cmd.Run()
	
	// Combine stdout and stderr for output
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	if err != nil {
		return BuildAppResult{
			Success: false,
			Message: fmt.Sprintf("build failed: %v", err),
			Output:  output,
		}, err
	}

	duration := time.Since(startTime)
	
	return BuildAppResult{
		Success: true,
		Message: fmt.Sprintf("Successfully built app '%s' in %s using %s", params.AppName, duration.Round(time.Millisecond), pm),
		Output:  output,
	}, nil
}

// CLI
func BuildAppCli() error {
	args := os.Args[3:]

	if len(args) != 1 {
		fmt.Println("Usage: layered-code tool build_app <app_name>")
		fmt.Println()
		fmt.Println("Build a Vite application for production")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  layered-code tool build_app myapp")
		return fmt.Errorf("build_app requires exactly 1 argument")
	}

	appName := args[0]
	result, err := BuildApp(BuildAppParams{AppName: appName})
	
	// Always print the output if available
	if result.Output != "" {
		fmt.Println(result.Output)
	}

	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("\n%s\n", result.Message)
	return nil
}

// MCP
func BuildAppMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	params := BuildAppParams{AppName: args.AppName}
	result, err := BuildApp(params)
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