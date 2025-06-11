package tools

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

// PM2Command represents available PM2 commands
type PM2Command string

const (
	PM2Start   PM2Command = "start"
	PM2Stop    PM2Command = "stop"
	PM2Restart PM2Command = "restart"
	PM2Delete  PM2Command = "delete"
	PM2Status  PM2Command = "status"
	PM2List    PM2Command = "list"
)

// PM2Params represents the parameters for PM2 commands
type PM2Params struct {
	AppName string     `json:"app_name"`
	Command PM2Command `json:"command"`
	Config  string     `json:"config,omitempty"` // Optional: config file (e.g., ecosystem.config.cjs)
}

// PM2Result represents the result of PM2 command
type PM2Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

// detectPM2PackageManager detects which package manager to use for PM2
func detectPM2PackageManager(appPath string) string {
	// Check for pnpm first
	if isPackageManagerAvailable("pnpm") {
		// Check if pnpm lockfile exists
		if _, err := os.Stat(filepath.Join(appPath, "pnpm-lock.yaml")); err == nil {
			return "pnpm"
		}
	}

	// Default to npm
	return "npm"
}

// runPM2Command executes a PM2 command
func runPM2Command(appPath string, command PM2Command, args ...string) (string, error) {
	// Detect package manager
	pm := detectPM2PackageManager(appPath)

	// Build command
	cmdArgs := []string{"pm2", string(command)}
	cmdArgs = append(cmdArgs, args...)

	// Use npx/pnpm dlx to run PM2
	var cmd *exec.Cmd
	if pm == "pnpm" {
		cmd = exec.Command("pnpm", append([]string{"dlx"}, cmdArgs...)...)
	} else {
		cmd = exec.Command("npx", cmdArgs...)
	}
	cmd.Dir = appPath

	// Set up output buffers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return stdout.String(), fmt.Errorf("%s", errorMsg)
	}

	return stdout.String(), nil
}

// PM2 executes PM2 commands
func PM2(params PM2Params) (PM2Result, error) {
	// Get the app directory
	appsDir, err := config.GetAppsDirectory()
	if err != nil {
		return PM2Result{Success: false, Message: fmt.Sprintf("failed to get apps directory: %v", err)}, err
	}

	appPath := filepath.Join(appsDir, params.AppName)

	// Check if app exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return PM2Result{Success: false, Message: fmt.Sprintf("app '%s' does not exist", params.AppName)}, fmt.Errorf("app not found")
	}

	// Handle different commands
	var output string
	var cmdErr error

	switch params.Command {
	case PM2Start:
		// If config file specified, use it
		if params.Config != "" {
			configPath := filepath.Join(appPath, params.Config)
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Try common config file names
				for _, configName := range []string{"ecosystem.config.cjs", "ecosystem.config.js", "ecosystem.dev.cjs", "ecosystem.dev.js"} {
					testPath := filepath.Join(appPath, configName)
					if _, err := os.Stat(testPath); err == nil {
						params.Config = configName
						break
					}
				}
				if params.Config == "" {
					return PM2Result{Success: false, Message: fmt.Sprintf("PM2 config file not found: %s", params.Config)}, fmt.Errorf("config file not found")
				}
			}
			output, cmdErr = runPM2Command(appPath, PM2Start, params.Config)
		} else {
			// Look for default config files
			configFound := false
			for _, configName := range []string{"ecosystem.config.cjs", "ecosystem.config.js", "ecosystem.dev.cjs", "ecosystem.dev.js"} {
				configPath := filepath.Join(appPath, configName)
				if _, err := os.Stat(configPath); err == nil {
					output, cmdErr = runPM2Command(appPath, PM2Start, configName)
					configFound = true
					break
				}
			}
			if !configFound {
				return PM2Result{Success: false, Message: "No PM2 config file found (ecosystem.config.cjs or similar)"}, fmt.Errorf("no config file found")
			}
		}

	case PM2Stop:
		output, cmdErr = runPM2Command(appPath, PM2Stop, "all")

	case PM2Restart:
		output, cmdErr = runPM2Command(appPath, PM2Restart, "all")

	case PM2Delete:
		output, cmdErr = runPM2Command(appPath, PM2Delete, "all")

	case PM2Status, PM2List:
		output, cmdErr = runPM2Command(appPath, PM2List, "--no-color")

	default:
		return PM2Result{Success: false, Message: fmt.Sprintf("unknown command: %s", params.Command)}, fmt.Errorf("unknown command")
	}

	if cmdErr != nil {
		return PM2Result{
			Success: false,
			Message: fmt.Sprintf("Failed to execute PM2 %s: %v", params.Command, cmdErr),
			Output:  output,
		}, cmdErr
	}

	return PM2Result{
		Success: true,
		Message: fmt.Sprintf("Successfully executed PM2 %s", params.Command),
		Output:  output,
	}, nil
}

// CLI
func PM2Cli() error {
	args := os.Args[3:]

	if len(args) < 2 {
		return fmt.Errorf("pm2 requires at least 2 arguments\nUsage: layered-code tool pm2 <app_name> <command> [config_file]\nCommands: start, stop, restart, delete, status")
	}

	appName := args[0]
	command := PM2Command(args[1])

	params := PM2Params{
		AppName: appName,
		Command: command,
	}

	// Check if config file is specified for start command
	if command == PM2Start && len(args) > 2 {
		params.Config = args[2]
	}

	// Validate command
	validCommands := []PM2Command{PM2Start, PM2Stop, PM2Restart, PM2Delete, PM2Status, PM2List}
	isValid := false
	for _, validCmd := range validCommands {
		if command == validCmd {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid command: %s. Valid commands: start, stop, restart, delete, status", command)
	}

	fmt.Printf("Executing PM2 %s for app '%s'...\n", command, appName)

	result, err := PM2(params)
	if err != nil {
		return fmt.Errorf("failed to execute PM2 command: %w", err)
	}

	fmt.Println(result.Message)
	if result.Output != "" {
		fmt.Println(result.Output)
	}

	return nil
}

// MCP
func PM2Mcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName string `json:"app_name"`
		Command string `json:"command"`
		Config  string `json:"config,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	params := PM2Params{
		AppName: args.AppName,
		Command: PM2Command(args.Command),
		Config:  args.Config,
	}

	result, err := PM2(params)
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
