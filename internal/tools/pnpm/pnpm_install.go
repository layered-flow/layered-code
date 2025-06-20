package pnpm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type PnpmInstallResult struct {
	AppName        string `json:"app_name"`
	AppPath        string `json:"app_path"`
	PackageManager string `json:"package_manager"`
	Message        string `json:"message"`
}

// PnpmInstall installs dependencies in an app directory using pnpm or npm
func PnpmInstall(appName string, showOutput bool) (PnpmInstallResult, error) {
	// Validate app name
	if appName == "" {
		return PnpmInstallResult{}, fmt.Errorf("app name is required")
	}

	// Get apps directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return PnpmInstallResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	// Create full app path
	appPath := filepath.Join(appsDir, appName)

	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return PnpmInstallResult{}, fmt.Errorf("app '%s' does not exist", appName)
	}

	// Check if package.json exists
	packageJsonPath := filepath.Join(appPath, "package.json")
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		return PnpmInstallResult{}, fmt.Errorf("package.json not found in app '%s'", appName)
	}

	// Determine package manager
	packageManager := "npm"
	if _, err := exec.LookPath("pnpm"); err == nil {
		packageManager = "pnpm"
	} else if _, err := exec.LookPath("npm"); err != nil {
		return PnpmInstallResult{}, fmt.Errorf("neither pnpm nor npm is available. Please install Node.js and npm or pnpm")
	}

	// Run install command
	cmd := exec.Command(packageManager, "install")
	cmd.Dir = appPath
	
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
		return PnpmInstallResult{}, fmt.Errorf("failed to install dependencies: %w", err)
	}

	return PnpmInstallResult{
		AppName:        appName,
		AppPath:        appPath,
		PackageManager: packageManager,
		Message:        fmt.Sprintf("Successfully installed dependencies for '%s' using %s", appName, packageManager),
	}, nil
}

// CLI
func PnpmInstallCli() error {
	args := os.Args[3:]

	if len(args) != 1 {
		return fmt.Errorf("usage: layered-code tool pnpm_install <app_name>")
	}

	appName := args[0]
	result, err := PnpmInstall(appName, true) // showOutput = true for CLI
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	fmt.Printf("\n%s\n", result.Message)
	fmt.Printf("Location: %s\n", result.AppPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", result.AppPath)
	fmt.Printf("  %s run dev\n", result.PackageManager)

	return nil
}

// MCP
func PnpmInstallMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := PnpmInstall(args.AppName, false) // showOutput = false for MCP
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