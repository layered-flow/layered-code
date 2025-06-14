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
		if isPackageManagerAvailable("npm") {
			return NPM
		}
	}

	// Default priority: pnpm -> npm
	if isPackageManagerAvailable("pnpm") {
		return PNPM
	}

	// Check if npm is available before returning it as fallback
	if isPackageManagerAvailable("npm") {
		return NPM
	}

	// If no package manager is available, return empty string
	// This will be handled with a proper error message
	return ""
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

	// Validate app name
	if params.AppName == "" {
		err := fmt.Errorf("app name cannot be empty")
		return NpmInstallResult{Success: false, Message: err.Error()}, err
	}
	if strings.ContainsAny(params.AppName, "/\\:*?\"<>|") || strings.Contains(params.AppName, "..") {
		err := fmt.Errorf("app name contains invalid characters")
		return NpmInstallResult{Success: false, Message: err.Error()}, err
	}

	// Get apps directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return NpmInstallResult{Success: false, Message: err.Error()}, err
	}

	// Build app path
	appPath := filepath.Join(appsDir, params.AppName)

	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		err = fmt.Errorf("app '%s' does not exist", params.AppName)
		return NpmInstallResult{Success: false, Message: err.Error()}, err
	} else if err != nil {
		err = fmt.Errorf("failed to check app existence: %w", err)
		return NpmInstallResult{Success: false, Message: err.Error()}, err
	}

	// Check if package.json exists
	packageJsonPath := filepath.Join(appPath, "package.json")
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		err = fmt.Errorf("package.json not found in app '%s'", params.AppName)
		return NpmInstallResult{Success: false, Message: err.Error()}, err
	} else if err != nil {
		err = fmt.Errorf("failed to check package.json: %w", err)
		return NpmInstallResult{Success: false, Message: err.Error()}, err
	}

	// Detect package manager
	pm := detectPackageManager(appPath, params.PackageManager)
	
	// Check if a package manager was found
	if pm == "" {
		return NpmInstallResult{
			Success: false, 
			Message: "No package manager found. Please install npm or pnpm. Visit https://nodejs.org to install Node.js and npm.",
		}, fmt.Errorf("no package manager available")
	}

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
		fmt.Println("Error: Invalid number of arguments")
		fmt.Println()
		fmt.Println("Usage: layered-code tool npm_install <app_name> [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  --package-manager=<pm>  - Force specific package manager (pnpm|npm|yarn)")
		fmt.Println("  --production           - Install only production dependencies")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  layered-code tool npm_install myapp")
		fmt.Println("  layered-code tool npm_install myapp --package-manager=pnpm")
		fmt.Println("  layered-code tool npm_install myapp --production")
		return fmt.Errorf("invalid arguments")
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
		// Print the user-friendly message from the result if available
		if result.Message != "" {
			fmt.Printf("Error: %s\n", result.Message)
		}
		return err
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