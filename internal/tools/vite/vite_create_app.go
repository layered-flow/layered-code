package vite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/helpers"
	"github.com/mark3labs/mcp-go/mcp"
)

// Types
type ViteCreateAppResult struct {
	AppName     string `json:"app_name"`
	AppPath     string `json:"app_path"`
	Template    string `json:"template"`
	Manager     string `json:"package_manager"`
	Message     string `json:"message"`
	ErrorOutput string `json:"error_output,omitempty"`
}

// ViteCreateApp creates a new Vite app in the apps directory with the specified template
func ViteCreateApp(appName string, template string, showOutput bool) (ViteCreateAppResult, error) {
	// Validate app name
	if err := helpers.ValidateAppName(appName); err != nil {
		return ViteCreateAppResult{}, err
	}

	// Validate template
	if template == "" {
		template = "react-ts" // Default to react-ts for TypeScript support
	}

	// List of valid templates
	validTemplates := map[string]bool{
		"vanilla": true, "vanilla-ts": true,
		"vue": true, "vue-ts": true,
		"react": true, "react-ts": true,
		"react-swc": true, "react-swc-ts": true,
		"preact": true, "preact-ts": true,
		"lit": true, "lit-ts": true,
		"svelte": true, "svelte-ts": true,
		"solid": true, "solid-ts": true,
		"qwik": true, "qwik-ts": true,
	}

	if !validTemplates[template] {
		return ViteCreateAppResult{}, fmt.Errorf("invalid template '%s'. Valid templates are: vanilla, vanilla-ts, vue, vue-ts, react, react-ts, react-swc, react-swc-ts, preact, preact-ts, lit, lit-ts, svelte, svelte-ts, solid, solid-ts, qwik, qwik-ts", template)
	}


	// Ensure apps directory exists
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return ViteCreateAppResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Create full app path
	appPath := filepath.Join(appsDir, appName)

	// Check if app already exists
	if _, err := os.Stat(appPath); err == nil {
		return ViteCreateAppResult{}, fmt.Errorf("app '%s' already exists", appName)
	}

	// Determine package manager
	packageManager := "npm"
	if _, err := exec.LookPath("pnpm"); err == nil {
		packageManager = "pnpm"
	} else if _, err := exec.LookPath("npm"); err != nil {
		return ViteCreateAppResult{}, fmt.Errorf("neither pnpm nor npm is available. Please install Node.js and npm or pnpm")
	}

	// Create the Vite app
	var cmd *exec.Cmd
	if packageManager == "pnpm" {
		cmd = exec.Command("pnpm", "create", "vite", appName, "--template", template, "--", "--yes")
	} else {
		cmd = exec.Command("npm", "create", "vite@latest", appName, "--", "--template", template)
	}
	
	cmd.Dir = appsDir
	
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
		// Clean up if creation failed
		os.RemoveAll(appPath)
		return ViteCreateAppResult{}, fmt.Errorf("failed to create Vite app: %w\nError output: %s", err, errBuf.String())
	}

	return ViteCreateAppResult{
		AppName:     appName,
		AppPath:     appPath,
		Template:    template,
		Manager:     packageManager,
		Message:     fmt.Sprintf("Successfully created Vite %s app '%s'. Run 'pnpm_install' or 'npm install' to install dependencies", template, appName),
		ErrorOutput: errBuf.String(),
	}, nil
}

// CLI
func ViteCreateAppCli() error {
	args := os.Args[3:]

	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: layered-code tool vite_create_app <app_name> [template]")
	}

	appName := args[0]
	template := "react-ts" // Default template
	if len(args) == 2 {
		template = args[1]
	}
	result, err := ViteCreateApp(appName, template, true) // showOutput = true for CLI
	if err != nil {
		return fmt.Errorf("failed to create Vite app: %w", err)
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
func ViteCreateAppMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName  string `json:"app_name"`
		Template string `json:"template,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := ViteCreateApp(args.AppName, args.Template, false) // showOutput = false for MCP
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