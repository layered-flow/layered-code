package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// PackageManager represents the available package managers
type PackageManager string

const (
	PNPM PackageManager = "pnpm"
	NPM  PackageManager = "npm"
	YARN PackageManager = "yarn"
)

// NpmInstallParams represents the parameters for npm install
type NpmInstallParams struct {
	AppName        string `json:"app_name"`
	PackageManager string `json:"package_manager,omitempty"` // Optional: force specific package manager
	Production     bool   `json:"production,omitempty"`      // Optional: install only production deps
}

// NpmInstallResult represents the result of npm install
type NpmInstallResult struct {
	Success        bool           `json:"success"`
	Message        string         `json:"message"`
	PackageManager PackageManager `json:"package_manager"`
	Duration       string         `json:"duration"`
}

// detectPackageManager detects which package manager to use
func detectPackageManager(appPath string, preferred string) PackageManager {
	// If user specified a preference, try that first
	if preferred != "" {
		pm := PackageManager(preferred)
		if isPackageManagerAvailable(string(pm)) {
			return pm
		}
	}

	// Check for existing lockfiles in the project
	if _, err := os.Stat(filepath.Join(appPath, "pnpm-lock.yaml")); err == nil {
		if isPackageManagerAvailable("pnpm") {
			return PNPM
		}
	}

	if _, err := os.Stat(filepath.Join(appPath, "yarn.lock")); err == nil {
		if isPackageManagerAvailable("yarn") {
			return YARN
		}
	}

	if _, err := os.Stat(filepath.Join(appPath, "package-lock.json")); err == nil {
		return NPM // npm is always available
	}

	// Default priority: pnpm -> npm
	if isPackageManagerAvailable("pnpm") {
		return PNPM
	}

	return NPM // npm is the universal fallback
}

// isPackageManagerAvailable checks if a package manager is available
func isPackageManagerAvailable(pm string) bool {
	cmd := exec.Command(pm, "--version")
	err := cmd.Run()
	return err == nil
}

// getInstallCommand returns the install command for the package manager
func getInstallCommand(pm PackageManager, production bool) []string {
	baseCmd := []string{string(pm), "install"}
	
	if production {
		switch pm {
		case PNPM:
			baseCmd = append(baseCmd, "--prod")
		case YARN:
			baseCmd = append(baseCmd, "--production")
		case NPM:
			baseCmd = append(baseCmd, "--production")
		}
	}
	
	return baseCmd
}

// NpmInstall installs npm dependencies
func NpmInstall(params NpmInstallParams) (NpmInstallResult, error) {
	startTime := time.Now()

	// Get the app directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return NpmInstallResult{Success: false, Message: fmt.Sprintf("failed to get apps directory: %v", err)}, err
	}

	appPath := filepath.Join(appsDir, params.AppName)

	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return NpmInstallResult{Success: false, Message: fmt.Sprintf("app '%s' does not exist", params.AppName)}, fmt.Errorf("app not found")
	}

	// Check if package.json exists
	packageJsonPath := filepath.Join(appPath, "package.json")
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		return NpmInstallResult{Success: false, Message: "No package.json found in app directory"}, fmt.Errorf("package.json not found")
	}

	// Detect package manager
	pm := detectPackageManager(appPath, params.PackageManager)

	// Get install command
	cmdArgs := getInstallCommand(pm, params.Production)

	// Create command
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = appPath

	// Set up output buffers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment variables for better output
	cmd.Env = append(os.Environ(),
		"FORCE_COLOR=0", // Disable color output for cleaner logs
	)

	// Run the command
	err = cmd.Run()
	
	duration := time.Since(startTime).Round(time.Second).String()

	if err != nil {
		// If pnpm failed and we haven't tried npm yet, fall back to npm
		if pm == PNPM && params.PackageManager == "" {
			// Try npm as fallback
			params.PackageManager = string(NPM)
			return NpmInstall(params)
		}

		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return NpmInstallResult{
			Success:        false,
			Message:        fmt.Sprintf("Failed to install dependencies: %s", errorMsg),
			PackageManager: pm,
			Duration:       duration,
		}, err
	}

	// Check if node_modules was created
	nodeModulesPath := filepath.Join(appPath, "node_modules")
	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		// For pnpm, this might be normal due to its linking strategy
		if pm != PNPM {
			return NpmInstallResult{
				Success:        false,
				Message:        "Installation completed but node_modules directory was not created",
				PackageManager: pm,
				Duration:       duration,
			}, fmt.Errorf("node_modules not created")
		}
	}

	return NpmInstallResult{
		Success:        true,
		Message:        fmt.Sprintf("Successfully installed dependencies using %s", pm),
		PackageManager: pm,
		Duration:       duration,
	}, nil
}

// CLI
func NpmInstallCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("npm_install requires at least 1 argument\nUsage: layered-code tool npm_install <app_name> [--package-manager=pnpm|npm|yarn] [--production]")
	}

	appName := args[0]
	params := NpmInstallParams{AppName: appName}

	// Parse additional arguments
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--package-manager=") {
			params.PackageManager = strings.TrimPrefix(arg, "--package-manager=")
		} else if arg == "--production" {
			params.Production = true
		}
	}

	fmt.Printf("Installing dependencies for app '%s'...\n", appName)

	result, err := NpmInstall(params)
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	fmt.Printf("%s (using %s, took %s)\n", result.Message, result.PackageManager, result.Duration)

	return nil
}

// MCP
func NpmInstallMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName        string `json:"app_name"`
		PackageManager string `json:"package_manager,omitempty"`
		Production     bool   `json:"production,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	params := NpmInstallParams{
		AppName:        args.AppName,
		PackageManager: args.PackageManager,
		Production:     args.Production,
	}

	result, err := NpmInstall(params)
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