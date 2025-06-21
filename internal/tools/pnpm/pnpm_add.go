package pnpm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/helpers"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type PnpmAddResult struct {
	AppName        string `json:"app_name"`
	AppPath        string `json:"app_path"`
	PackageManager string `json:"package_manager"`
	Package        string `json:"package"`
	Message        string `json:"message"`
	Output         string `json:"output,omitempty"`
	ErrorOutput    string `json:"error_output,omitempty"`
}

// PnpmAdd adds a package to an app directory using pnpm or npm
func PnpmAdd(appName string, packageName string, showOutput bool) (PnpmAddResult, error) {
	// Validate app name
	if appName == "" {
		return PnpmAddResult{}, fmt.Errorf("app name is required")
	}
	
	if err := helpers.ValidateAppName(appName); err != nil {
		return PnpmAddResult{}, fmt.Errorf("invalid app name: %w", err)
	}

	// Validate package name
	if packageName == "" {
		return PnpmAddResult{}, fmt.Errorf("package name is required")
	}

	// Get apps directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return PnpmAddResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	// Create full app path
	appPath := filepath.Join(appsDir, appName)

	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return PnpmAddResult{}, fmt.Errorf("app '%s' does not exist", appName)
	}

	// Check if package.json exists
	packageJsonPath := filepath.Join(appPath, "package.json")
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		return PnpmAddResult{}, fmt.Errorf("package.json not found in app '%s'", appName)
	}

	// Determine package manager
	packageManager, err := DetectPackageManager()
	if err != nil {
		return PnpmAddResult{}, err
	}

	// Build command
	var cmd *exec.Cmd
	if packageManager == "pnpm" {
		cmd = exec.Command("pnpm", "add", packageName)
	} else {
		cmd = exec.Command("npm", "install", packageName)
	}
	cmd.Dir = appPath
	
	// Capture output
	var outBuf, errBuf bytes.Buffer
	
	if showOutput {
		// For CLI, use MultiWriter to both stream and capture output
		cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)
	} else {
		// For MCP, just capture to buffers
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
	}

	if err := cmd.Run(); err != nil {
		return PnpmAddResult{}, fmt.Errorf("failed to add package '%s': %w\nError output: %s", packageName, err, errBuf.String())
	}

	return PnpmAddResult{
		AppName:        appName,
		AppPath:        appPath,
		PackageManager: packageManager,
		Package:        packageName,
		Message:        fmt.Sprintf("Successfully added '%s' to '%s' using %s", packageName, appName, packageManager),
		Output:         outBuf.String(),
		ErrorOutput:    errBuf.String(),
	}, nil
}

// CLI
func PnpmAddCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("usage: layered-code tool pnpm_add <app_name> <package_name>")
	}

	appName := args[0]
	packageName := strings.Join(args[1:], " ") // Join remaining args to support scoped packages
	
	result, err := PnpmAdd(appName, packageName, true) // showOutput = true for CLI
	if err != nil {
		return fmt.Errorf("failed to add package: %w", err)
	}

	fmt.Printf("\n%s\n", result.Message)
	fmt.Printf("Location: %s\n", result.AppPath)

	return nil
}

// MCP
func PnpmAddMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName     string `json:"app_name"`
		PackageName string `json:"package_name"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := PnpmAdd(args.AppName, args.PackageName, false) // showOutput = false for MCP
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