package vite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type ViteCreateReactAppResult struct {
	AppName   string `json:"app_name"`
	AppPath   string `json:"app_path"`
	Manager   string `json:"package_manager"`
	Message   string `json:"message"`
}

// ViteCreateReactApp creates a new Vite React app in the apps directory
func ViteCreateReactApp(appName string, showOutput bool) (ViteCreateReactAppResult, error) {
	// Validate app name
	if appName == "" {
		return ViteCreateReactAppResult{}, fmt.Errorf("app name is required")
	}

	// Check for directory traversal attempts
	if strings.Contains(appName, "..") || strings.Contains(appName, "/") || strings.Contains(appName, "\\") {
		return ViteCreateReactAppResult{}, fmt.Errorf("app name cannot contain path separators or '..'")
	}

	// Ensure apps directory exists
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return ViteCreateReactAppResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Create full app path
	appPath := filepath.Join(appsDir, appName)

	// Check if app already exists
	if _, err := os.Stat(appPath); err == nil {
		return ViteCreateReactAppResult{}, fmt.Errorf("app '%s' already exists", appName)
	}

	// Determine package manager
	packageManager := "npm"
	if _, err := exec.LookPath("pnpm"); err == nil {
		packageManager = "pnpm"
	} else if _, err := exec.LookPath("npm"); err != nil {
		return ViteCreateReactAppResult{}, fmt.Errorf("neither pnpm nor npm is available. Please install Node.js and npm or pnpm")
	}

	// Create the Vite app
	var cmd *exec.Cmd
	if packageManager == "pnpm" {
		cmd = exec.Command("pnpm", "create", "vite", appName, "--template", "react", "--", "--yes")
	} else {
		cmd = exec.Command("npm", "create", "vite@latest", appName, "--", "--template", "react")
	}
	
	cmd.Dir = appsDir
	
	// If showOutput is true (CLI), stream to stdout/stderr
	// If false (MCP), capture the output
	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Capture output for MCP to avoid polluting JSON response
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
	}

	if err := cmd.Run(); err != nil {
		// Clean up if creation failed
		os.RemoveAll(appPath)
		return ViteCreateReactAppResult{}, fmt.Errorf("failed to create Vite app: %w", err)
	}

	return ViteCreateReactAppResult{
		AppName: appName,
		AppPath: appPath,
		Manager: packageManager,
		Message: fmt.Sprintf("Successfully created Vite React app '%s'. Run 'pnpm_install' or 'npm install' to install dependencies", appName),
	}, nil
}

// CLI
func ViteCreateReactAppCli() error {
	args := os.Args[3:]

	if len(args) != 1 {
		return fmt.Errorf("usage: layered-code tool vite_create_react_app <app_name>")
	}

	appName := args[0]
	result, err := ViteCreateReactApp(appName, true) // showOutput = true for CLI
	if err != nil {
		return fmt.Errorf("failed to create Vite React app: %w", err)
	}

	fmt.Printf("\n%s\n", result.Message)
	fmt.Printf("Location: %s\n", result.AppPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  layered-code tool pnpm_install %s\n", result.AppName)
	fmt.Printf("  cd %s\n", result.AppPath)
	fmt.Printf("  %s run dev\n", result.Manager)

	return nil
}

// MCP
func ViteCreateReactAppMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := ViteCreateReactApp(args.AppName, false) // showOutput = false for MCP
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