package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/layered-flow/layered-code/internal/notifications"
	"github.com/mark3labs/mcp-go/mcp"
)

// WriteFileParams represents the parameters for writing a file
type WriteFileParams struct {
	AppName  string `json:"app_name"`
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	Mode     string `json:"mode"` // "create" or "overwrite"
}

// WriteFileResult represents the result of writing a file
type WriteFileResult struct {
	AppName      string     `json:"app_name"`
	FilePath     string     `json:"file_path"`
	BytesWritten int        `json:"bytes_written"`
	Created      bool       `json:"created"`
	LastModified *time.Time `json:"last_modified,omitempty"`
}

// WriteFile writes content to a file within an app directory
func WriteFile(params WriteFileParams) (WriteFileResult, error) {
	if params.AppName == "" {
		return WriteFileResult{}, errors.New("app_name is required")
	}
	if params.FilePath == "" {
		return WriteFileResult{}, errors.New("file_path is required")
	}
	if params.Mode == "" {
		params.Mode = "create" // Default mode
	}
	if params.Mode != "create" && params.Mode != "overwrite" {
		return WriteFileResult{}, fmt.Errorf("invalid mode: %s (must be 'create' or 'overwrite')", params.Mode)
	}

	// Check file size limit
	if len(params.Content) > int(constants.MaxFileSize) {
		return WriteFileResult{}, fmt.Errorf("content exceeds maximum file size of %s", constants.MaxFileSizeInWords)
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return WriteFileResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct app and output directories
	appDir := filepath.Join(appsDir, params.AppName)
	outputDir := filepath.Join(appDir, constants.OutputDirectoryName)
	
	// Construct and validate the full file path
	fullPath := filepath.Join(outputDir, params.FilePath)
	cleanPath := filepath.Clean(fullPath)

	// Ensure the file is within the build directory
	if !config.IsWithinDirectory(cleanPath, outputDir) {
		return WriteFileResult{}, fmt.Errorf("file path attempts to access file outside build directory")
	}

	// Check if app directory exists
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return WriteFileResult{}, fmt.Errorf("app directory does not exist: %s", params.AppName)
	}
	
	// Ensure build directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return WriteFileResult{}, fmt.Errorf("failed to create build directory: %w", err)
	}

	// Check if file exists
	fileExists := false
	if info, err := os.Stat(cleanPath); err == nil {
		if info.IsDir() {
			return WriteFileResult{}, fmt.Errorf("path is a directory, not a file")
		}
		fileExists = true
	}

	// Handle create vs overwrite mode
	if params.Mode == "create" && fileExists {
		return WriteFileResult{}, fmt.Errorf("file already exists (use mode 'overwrite' to replace)")
	}

	// Create parent directories if needed
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return WriteFileResult{}, fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Write the file
	if err := os.WriteFile(cleanPath, []byte(params.Content), 0644); err != nil {
		return WriteFileResult{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info for the result
	info, err := os.Stat(cleanPath)
	if err != nil {
		return WriteFileResult{}, fmt.Errorf("failed to stat written file: %w", err)
	}

	// Send WebSocket notification
	action := "write"
	if !fileExists {
		action = "create"
	}
	notificationPath := filepath.Join(params.AppName, params.FilePath)
	notifications.NotifyFileChange(notificationPath, action)

	modTime := info.ModTime()
	return WriteFileResult{
		AppName:      params.AppName,
		FilePath:     params.FilePath,
		BytesWritten: len(params.Content),
		Created:      !fileExists,
		LastModified: &modTime,
	}, nil
}

// CLI
func WriteFileCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printWriteFileHelp()
			return nil
		}
	}

	var params WriteFileParams
	var contentFile string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--app-name":
			if i+1 < len(args) {
				params.AppName = args[i+1]
				i++
			} else {
				return errors.New("--app-name requires a value")
			}
		case "--file-path":
			if i+1 < len(args) {
				params.FilePath = args[i+1]
				i++
			} else {
				return errors.New("--file-path requires a value")
			}
		case "--content":
			if i+1 < len(args) {
				params.Content = args[i+1]
				i++
			} else {
				return errors.New("--content requires a value")
			}
		case "--content-file":
			if i+1 < len(args) {
				contentFile = args[i+1]
				i++
			} else {
				return errors.New("--content-file requires a value")
			}
		case "--mode":
			if i+1 < len(args) {
				params.Mode = args[i+1]
				i++
			} else {
				return errors.New("--mode requires a value")
			}
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool write_file --help' for usage", args[i])
			}
		}
	}

	if params.AppName == "" {
		return errors.New("--app-name is required")
	}
	if params.FilePath == "" {
		return errors.New("--file-path is required")
	}

	// Handle content source
	if params.Content == "" && contentFile == "" {
		return errors.New("either --content or --content-file is required")
	}
	if params.Content != "" && contentFile != "" {
		return errors.New("cannot use both --content and --content-file")
	}

	// Read content from file if specified
	if contentFile != "" {
		content, err := os.ReadFile(contentFile)
		if err != nil {
			return fmt.Errorf("failed to read content file: %w", err)
		}
		params.Content = string(content)
	}

	result, err := WriteFile(params)
	if err != nil {
		return err
	}

	action := "Updated"
	if result.Created {
		action = "Created"
	}
	fmt.Printf("%s file: %s/%s\n", action, result.AppName, result.FilePath)
	fmt.Printf("Bytes written: %d\n", result.BytesWritten)
	return nil
}

func printWriteFileHelp() {
	fmt.Println("Usage: layered-code tool write_file [options]")
	fmt.Println()
	fmt.Println("Write or create a file within an application directory")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>    Name of the app directory")
	fmt.Println("  --file-path <path>   Path to the file relative to the app directory")
	fmt.Println()
	fmt.Println("Content options (one required):")
	fmt.Println("  --content <text>     Content to write to the file")
	fmt.Println("  --content-file <path> Read content from the specified file")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  --mode <mode>        Write mode: 'create' (default) or 'overwrite'")
	fmt.Println("                       'create' fails if file exists")
	fmt.Println("                       'overwrite' replaces existing file")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Parent directories will be created automatically")
	fmt.Printf("  - Maximum file size is %s\n", constants.MaxFileSizeInWords)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create a new file")
	fmt.Println("  layered-code tool write_file --app-name myapp --file-path src/new.go --content 'package main'")
	fmt.Println()
	fmt.Println("  # Overwrite an existing file")
	fmt.Println("  layered-code tool write_file --app-name myapp --file-path config.json --mode overwrite --content '{}'")
	fmt.Println()
	fmt.Println("  # Write content from another file")
	fmt.Println("  layered-code tool write_file --app-name myapp --file-path data.txt --content-file /tmp/source.txt")
}

// MCP
func WriteFileMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params WriteFileParams

	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	result, err := WriteFile(params)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}