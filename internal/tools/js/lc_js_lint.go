package js

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/helpers"

	"github.com/mark3labs/mcp-go/mcp"
)

// shellEscape quotes a string for safe shell usage
func shellEscape(s string) string {
	// Use Go's built-in quoting which handles special characters properly
	return fmt.Sprintf("%q", s)
}

// Types
type LcJsLintParams struct {
	AppName string   `json:"app_name"`
	Files   []string `json:"files"`
	Config  string   `json:"config,omitempty"`
}

type LcJsLintResult struct {
	Success     bool   `json:"success"`
	Output      string `json:"output"`
	ExitCode    int    `json:"exit_code"`
	Command     string `json:"command"`
	Message     string `json:"message,omitempty"`
	ErrorOutput string `json:"error_output,omitempty"`
}

// LcJsLint runs ESLint analysis on JavaScript/TypeScript files
func LcJsLint(params LcJsLintParams) (LcJsLintResult, error) {
	// Validate app name
	if err := helpers.ValidateAppName(params.AppName); err != nil {
		return LcJsLintResult{}, fmt.Errorf("invalid app name: %w", err)
	}

	// Validate files parameter
	if len(params.Files) == 0 {
		return LcJsLintResult{}, fmt.Errorf("files parameter cannot be empty")
	}

	// Validate each file pattern
	for _, pattern := range params.Files {
		if pattern == "" {
			return LcJsLintResult{}, fmt.Errorf("file patterns cannot be empty")
		}
		// Validate against path traversal
		if strings.Contains(pattern, "..") {
			return LcJsLintResult{}, fmt.Errorf("file pattern cannot contain '..': %s", pattern)
		}
	}

	// Get apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcJsLintResult{}, fmt.Errorf("failed to get apps directory: %w", err)
	}

	// Build app directory path
	appDir := filepath.Join(appsDir, params.AppName)

	// Check if app directory exists
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return LcJsLintResult{
			Success:     false,
			Message:     fmt.Sprintf("App directory does not exist: %s", params.AppName),
			ErrorOutput: fmt.Sprintf("Directory not found: %s", appDir),
		}, nil
	}

	// Build eslint command
	args := []string{"--format", "json"}

	// Add config file if specified
	if params.Config != "" {
		// Validate that config path doesn't contain directory traversal
		if strings.Contains(params.Config, "..") {
			return LcJsLintResult{}, fmt.Errorf("config path cannot contain '..'")
		}
		configPath := filepath.Join(appDir, params.Config)
		args = append(args, "--config", configPath)
	}

	// Add files to lint
	args = append(args, params.Files...)

	// Create command
	cmd := exec.Command("npx", append([]string{"eslint"}, args...)...)
	cmd.Dir = appDir

	// Run the command
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			// ESLint returns exit code 1 when there are linting errors, which is expected
			if exitCode != 1 {
				// Build properly escaped command string
				escapedArgs := make([]string, len(args))
				for i, arg := range args {
					escapedArgs[i] = shellEscape(arg)
				}
				return LcJsLintResult{
					Success:     false,
					ExitCode:    exitCode,
					Output:      string(output),
					Command:     fmt.Sprintf("cd %s && npx eslint %s", shellEscape(appDir), strings.Join(escapedArgs, " ")),
					Message:     fmt.Sprintf("ESLint command failed with exit code %d", exitCode),
					ErrorOutput: string(output),
				}, nil
			}
		} else {
			// Check if eslint is not found
			if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "command not found") {
				// Build properly escaped command string
				escapedArgs := make([]string, len(args))
				for i, arg := range args {
					escapedArgs[i] = shellEscape(arg)
				}
				return LcJsLintResult{
					Success:     false,
					Command:     fmt.Sprintf("cd %s && npx eslint %s", shellEscape(appDir), strings.Join(escapedArgs, " ")),
					Message:     "ESLint not found. Please install ESLint in your project: npm install --save-dev eslint",
					ErrorOutput: err.Error(),
				}, nil
			}
			// Build properly escaped command string
			escapedArgs := make([]string, len(args))
			for i, arg := range args {
				escapedArgs[i] = shellEscape(arg)
			}
			return LcJsLintResult{
				Success:     false,
				Command:     fmt.Sprintf("cd %s && npx eslint %s", shellEscape(appDir), strings.Join(escapedArgs, " ")),
				Message:     "Failed to run ESLint",
				ErrorOutput: err.Error(),
			}, nil
		}
	}

	// Build command string for display with proper escaping
	escapedArgs := make([]string, len(args))
	for i, arg := range args {
		escapedArgs[i] = shellEscape(arg)
	}
	commandStr := fmt.Sprintf("cd %s && npx eslint %s", shellEscape(appDir), strings.Join(escapedArgs, " "))

	return LcJsLintResult{
		Success:  exitCode == 0,
		Output:   string(output),
		ExitCode: exitCode,
		Command:  commandStr,
	}, nil
}

// CLI
func LcJsLintCli() error {
	args := os.Args[3:]

	if len(args) < 1 {
		return fmt.Errorf("usage: layered-code tool lc_js_lint <app_name> [--config=<config_file>] <files...>")
	}

	params := LcJsLintParams{
		AppName: args[0],
	}

	// Parse remaining arguments
	fileArgs := []string{}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--config=") {
			params.Config = strings.TrimPrefix(arg, "--config=")
		} else {
			fileArgs = append(fileArgs, arg)
		}
	}

	if len(fileArgs) == 0 {
		return fmt.Errorf("no files specified to lint")
	}

	params.Files = fileArgs

	result, err := LcJsLint(params)
	if err != nil {
		return fmt.Errorf("failed to run eslint: %w", err)
	}

	// Print command
	fmt.Printf("Command: %s\n\n", result.Command)

	// Print output
	if result.Output != "" {
		// Try to parse and pretty print JSON output
		var jsonData interface{}
		if err := json.Unmarshal([]byte(result.Output), &jsonData); err == nil {
			prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
			if err == nil {
				fmt.Println(string(prettyJSON))
			} else {
				fmt.Println(result.Output)
			}
		} else {
			fmt.Println(result.Output)
		}
	}

	// Print summary
	if result.Success {
		fmt.Println("\n✓ No linting errors found")
	} else {
		fmt.Printf("\n✗ ESLint found issues (exit code: %d)\n", result.ExitCode)
	}

	return nil
}

// MCP
func LcJsLintMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LcJsLintParams
	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	result, err := LcJsLint(params)
	if err != nil {
		// For MCP, we still return actual errors for parameter validation
		return nil, err
	}

	// Convert result to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

