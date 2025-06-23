package nextjs

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
type NextjsAppCreateResult struct {
	AppName     string `json:"app_name"`
	AppPath     string `json:"app_path"`
	Template    string `json:"template"`
	Manager     string `json:"package_manager"`
	Message     string `json:"message"`
	ErrorOutput string `json:"error_output,omitempty"`
}

// NextjsAppCreate creates a new Next.js app in the apps directory with the specified template
func NextjsAppCreate(appName string, template string, showOutput bool) (NextjsAppCreateResult, error) {
	// Validate app name
	if err := helpers.ValidateAppName(appName); err != nil {
		return NextjsAppCreateResult{}, err
	}

	// Validate template
	if template == "" {
		template = "typescript" // Default to TypeScript
	}

	// List of valid templates
	validTemplates := map[string]bool{
		"typescript":        true,
		"javascript":        true,
		"tailwind":         true,
		"app":              true, // App router with TypeScript
		"app-tw":           true, // App router with TypeScript and Tailwind
	}

	if !validTemplates[template] {
		return NextjsAppCreateResult{}, fmt.Errorf("invalid template '%s'. Valid templates are: typescript, javascript, tailwind, app, app-tw", template)
	}

	// Ensure apps directory exists
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return NextjsAppCreateResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Create full app path
	appPath := filepath.Join(appsDir, appName)

	// Check if app already exists
	if _, err := os.Stat(appPath); err == nil {
		return NextjsAppCreateResult{}, fmt.Errorf("app '%s' already exists", appName)
	}

	// Determine package manager
	packageManager := "npm"
	if _, err := exec.LookPath("pnpm"); err == nil {
		packageManager = "pnpm"
	} else if _, err := exec.LookPath("npm"); err != nil {
		return NextjsAppCreateResult{}, fmt.Errorf("neither pnpm nor npm is available. Please install Node.js and npm or pnpm")
	}

	// Build create-next-app command
	var cmd *exec.Cmd
	args := []string{}
	
	if packageManager == "pnpm" {
		args = append(args, "create", "next-app", appName)
	} else {
		args = append(args, "create", "next-app@latest", appName)
	}

	// Add template-specific flags
	switch template {
	case "typescript":
		args = append(args, "--typescript", "--no-tailwind", "--app", "--eslint", "--no-src-dir", "--import-alias", "@/*", "--turbopack")
	case "javascript":
		args = append(args, "--js", "--no-tailwind", "--app", "--eslint", "--no-src-dir", "--import-alias", "@/*", "--turbopack")
	case "tailwind":
		args = append(args, "--typescript", "--tailwind", "--app", "--eslint", "--no-src-dir", "--import-alias", "@/*", "--turbopack")
	case "app":
		args = append(args, "--typescript", "--no-tailwind", "--app", "--eslint", "--no-src-dir", "--import-alias", "@/*", "--turbopack")
	case "app-tw":
		args = append(args, "--typescript", "--tailwind", "--app", "--eslint", "--no-src-dir", "--import-alias", "@/*", "--turbopack")
	}

	// Use package manager preferences
	if packageManager == "pnpm" {
		args = append(args, "--use-pnpm")
	} else {
		args = append(args, "--use-npm")
	}

	// Skip install for now (user will run pnpm_install separately)
	args = append(args, "--skip-install")

	if packageManager == "pnpm" {
		cmd = exec.Command("pnpm", args...)
	} else {
		cmd = exec.Command("npm", args...)
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
		return NextjsAppCreateResult{}, fmt.Errorf("failed to create Next.js app: %w\nError output: %s", err, errBuf.String())
	}

	return NextjsAppCreateResult{
		AppName:     appName,
		AppPath:     appPath,
		Template:    template,
		Manager:     packageManager,
		Message:     fmt.Sprintf("Successfully created Next.js %s app '%s'. Run 'pnpm_install' or 'npm install' to install dependencies", template, appName),
		ErrorOutput: errBuf.String(),
	}, nil
}

// CLI
func NextjsAppCreateCli() error {
	args := os.Args[3:]

	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("usage: layered-code tool nextjs_app_create <app_name> [template]")
	}

	appName := args[0]
	template := "typescript" // Default template
	if len(args) == 2 {
		template = args[1]
	}
	result, err := NextjsAppCreate(appName, template, true) // showOutput = true for CLI
	if err != nil {
		return fmt.Errorf("failed to create Next.js app: %w", err)
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
func NextjsAppCreateMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName  string `json:"app_name"`
		Template string `json:"template,omitempty"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := NextjsAppCreate(args.AppName, args.Template, false) // showOutput = false for MCP
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