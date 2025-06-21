package pnpm

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
type PnpmPm2Result struct {
	AppName        string `json:"app_name,omitempty"`
	AppPath        string `json:"app_path,omitempty"`
	PackageManager string `json:"package_manager"`
	Command        string `json:"command"`
	Output         string `json:"output,omitempty"`
	Message        string `json:"message"`
}

// PackageJSON represents the structure of package.json
type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

// PnpmPm2 executes PM2 commands with opinionated defaults for app management
func PnpmPm2(command string, target string, showOutput bool) (PnpmPm2Result, error) {
	// Determine package manager
	packageManager, err := DetectPackageManager()
	if err != nil {
		return PnpmPm2Result{}, err
	}

	// Use pnpm dlx or npx to handle PM2
	var pm2Prefix string
	if packageManager == "pnpm" {
		pm2Prefix = "pnpm dlx pm2"
	} else {
		pm2Prefix = "npx pm2"
	}
	
	var pm2Command string
	var appPath string
	var appName string

	switch command {
	case "start":
		if target == "" {
			return PnpmPm2Result{}, fmt.Errorf("app name is required for start command")
		}
		
		// Get apps directory
		appsDir, err := config.GetAppsDirectory()
		if err != nil {
			return PnpmPm2Result{}, fmt.Errorf("failed to get apps directory: %w", err)
		}
		
		appName = target
		appPath = filepath.Join(appsDir, appName)
		
		// Check if app exists
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			return PnpmPm2Result{}, fmt.Errorf("app '%s' does not exist", appName)
		}
		
		// Check for ecosystem.config.js
		ecosystemPath := filepath.Join(appPath, "ecosystem.config.js")
		if _, err := os.Stat(ecosystemPath); err == nil {
			// Use ecosystem file
			pm2Command = fmt.Sprintf("%s start ecosystem.config.js", pm2Prefix)
		} else {
			// No ecosystem file, determine script to run
			packageJsonPath := filepath.Join(appPath, "package.json")
			scriptToRun, err := getScriptToRun(packageJsonPath)
			if err != nil {
				return PnpmPm2Result{}, fmt.Errorf("failed to determine script to run: %w", err)
			}
			
			// Build PM2 start command with the detected script
			pm2Command = fmt.Sprintf("%s start \"%s run %s\" --name %s", pm2Prefix, packageManager, scriptToRun, appName)
		}
		
	case "stop", "restart", "delete":
		if target == "" {
			return PnpmPm2Result{}, fmt.Errorf("target (app name or 'all') is required for %s command", command)
		}
		pm2Command = fmt.Sprintf("%s %s %s", pm2Prefix, command, target)
		
	case "list", "status", "ls":
		pm2Command = fmt.Sprintf("%s list", pm2Prefix)
		
	case "logs":
		if target != "" {
			pm2Command = fmt.Sprintf("%s logs %s", pm2Prefix, target)
		} else {
			pm2Command = fmt.Sprintf("%s logs", pm2Prefix)
		}
		
	default:
		return PnpmPm2Result{}, fmt.Errorf("unsupported PM2 command: %s. Supported commands: start, stop, restart, delete, list, logs", command)
	}
	
	// Execute the command
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	
	cmd := exec.Command(shell, "-l", "-c", pm2Command)
	if appPath != "" {
		cmd.Dir = appPath
	}
	
	// Capture output
	var outBuf, errBuf bytes.Buffer
	
	if showOutput {
		// For CLI, also stream to stdout/stderr
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
	}
	
	// Always capture for result
	if !showOutput {
		if err := cmd.Run(); err != nil {
			return PnpmPm2Result{}, fmt.Errorf("failed to execute pm2 command '%s': %w\nError output: %s", pm2Command, err, errBuf.String())
		}
	} else {
		// For CLI, we need to capture output too
		cmd = exec.Command(shell, "-l", "-c", pm2Command)
		if appPath != "" {
			cmd.Dir = appPath
		}
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		
		if err := cmd.Run(); err != nil {
			return PnpmPm2Result{}, fmt.Errorf("failed to execute pm2 command '%s': %w\nError output: %s", pm2Command, err, errBuf.String())
		}
		
		// Print the output
		if outBuf.Len() > 0 {
			fmt.Print(outBuf.String())
		}
		if errBuf.Len() > 0 {
			fmt.Fprint(os.Stderr, errBuf.String())
		}
	}
	
	result := PnpmPm2Result{
		PackageManager: packageManager,
		Command:        pm2Command,
		Output:         outBuf.String(),
		Message:        fmt.Sprintf("Successfully executed: %s", pm2Command),
	}
	
	if appName != "" {
		result.AppName = appName
		result.AppPath = appPath
	}
	
	return result, nil
}

// getScriptToRun reads package.json and determines which script to run
func getScriptToRun(packageJsonPath string) (string, error) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return "", fmt.Errorf("failed to read package.json: %w", err)
	}
	
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse package.json: %w", err)
	}
	
	// Priority order: dev, start, main/index.js
	if _, ok := pkg.Scripts["dev"]; ok {
		return "dev", nil
	}
	if _, ok := pkg.Scripts["start"]; ok {
		return "start", nil
	}
	
	// Check for main entry point files
	for _, file := range []string{"index.js", "main.js", "server.js", "app.js"} {
		if _, err := os.Stat(filepath.Join(filepath.Dir(packageJsonPath), file)); err == nil {
			return file, nil
		}
	}
	
	return "", fmt.Errorf("no suitable script found in package.json (looked for 'dev' or 'start' scripts)")
}

// CLI
func PnpmPm2Cli() error {
	args := os.Args[3:]
	
	if len(args) < 1 {
		return fmt.Errorf("usage: layered-code tool pnpm_pm2 <command> [target]\n" +
			"Commands:\n" +
			"  start <app-name>    Start an app (uses ecosystem.config.js or package.json scripts)\n" +
			"  stop <app-name|all> Stop a specific app or all apps\n" +
			"  restart <app-name|all> Restart a specific app or all apps\n" +
			"  delete <app-name|all>  Delete a specific app or all apps\n" +
			"  list                Show all PM2 processes\n" +
			"  logs [app-name]     Show logs for all apps or a specific app")
	}
	
	command := args[0]
	var target string
	if len(args) > 1 {
		target = args[1]
	}
	
	result, err := PnpmPm2(command, target, true)
	if err != nil {
		return fmt.Errorf("failed to execute pm2 command: %w", err)
	}
	
	// Output already printed by PnpmPm2 when showOutput=true
	// Just print the success message if not already shown
	if !strings.Contains(result.Output, result.Message) {
		fmt.Printf("\n%s\n", result.Message)
	}
	
	return nil
}

// MCP
func PnpmPm2Mcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		Command string `json:"command"`
		Target  string `json:"target,omitempty"`
	}
	
	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}
	
	if args.Command == "" {
		return nil, fmt.Errorf("command is required")
	}
	
	result, err := PnpmPm2(args.Command, args.Target, false)
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