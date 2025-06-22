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
	AppName        string   `json:"app_name"`
	AppPath        string   `json:"app_path"`
	PackageManager string   `json:"package_manager"`
	Packages       []string `json:"packages"`
	Message        string   `json:"message"`
	Output         string   `json:"output,omitempty"`
	ErrorOutput    string   `json:"error_output,omitempty"`
}

// parsePackageNames splits package names by spaces but preserves scoped packages
// e.g., "react @types/react react-dom" -> ["react", "@types/react", "react-dom"]
func parsePackageNames(packageNames string) []string {
	// Trim spaces and split by whitespace
	parts := strings.Fields(strings.TrimSpace(packageNames))
	
	var packages []string
	for _, part := range parts {
		if part != "" {
			packages = append(packages, part)
		}
	}
	
	return packages
}

// PnpmAdd adds one or more packages to an app directory using pnpm or npm
func PnpmAdd(appName string, packageNames string, showOutput bool) (PnpmAddResult, error) {
	// Validate app name
	if appName == "" {
		return PnpmAddResult{}, fmt.Errorf("app name is required")
	}
	
	if err := helpers.ValidateAppName(appName); err != nil {
		return PnpmAddResult{}, fmt.Errorf("invalid app name: %w", err)
	}

	// Validate package names
	if packageNames == "" {
		return PnpmAddResult{}, fmt.Errorf("package name(s) required")
	}

	// Split package names by spaces, but handle scoped packages correctly
	packages := parsePackageNames(packageNames)

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

	// Build command with all packages
	var cmd *exec.Cmd
	if packageManager == "pnpm" {
		args := append([]string{"add"}, packages...)
		cmd = exec.Command("pnpm", args...)
	} else {
		args := append([]string{"install"}, packages...)
		cmd = exec.Command("npm", args...)
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
		return PnpmAddResult{}, fmt.Errorf("failed to add package(s) '%s': %w\nError output: %s", strings.Join(packages, " "), err, errBuf.String())
	}

	// Create appropriate message based on number of packages
	var message string
	if len(packages) == 1 {
		message = fmt.Sprintf("Successfully added '%s' to '%s' using %s", packages[0], appName, packageManager)
	} else {
		message = fmt.Sprintf("Successfully added %d packages to '%s' using %s: %s", len(packages), appName, packageManager, strings.Join(packages, ", "))
	}

	return PnpmAddResult{
		AppName:        appName,
		AppPath:        appPath,
		PackageManager: packageManager,
		Packages:       packages,
		Message:        message,
		Output:         outBuf.String(),
		ErrorOutput:    errBuf.String(),
	}, nil
}

// CLI
func PnpmAddCli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("usage: layered-code tool pnpm_add <app_name> <package_name> [package_name...]")
	}

	appName := args[0]
	packageNames := strings.Join(args[1:], " ") // Join all package names
	
	result, err := PnpmAdd(appName, packageNames, true) // showOutput = true for CLI
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