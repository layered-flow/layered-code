package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	ErrSymlink      = errors.New("file is a symlink")
	ErrBinaryFile   = errors.New("file appears to be binary")
	ErrFileTooLarge = errors.New("file exceeds maximum size of " + constants.MaxFileSizeInWords)
)

// ReadFileResult represents the result of reading a file
type ReadFileResult struct {
	AppName  string `json:"app_name"`
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// ReadFile reads the content of a file within an app directory
func ReadFile(appName, filePath string) (ReadFileResult, error) {
	if appName == "" {
		return ReadFileResult{}, errors.New("app_name is required")
	}
	if filePath == "" {
		return ReadFileResult{}, errors.New("file_path is required")
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return ReadFileResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct and validate the full file path
	fullPath := filepath.Join(appsDir, appName, filePath)
	cleanPath := filepath.Clean(fullPath)

	// Ensure the file is within the app directory
	appDir := filepath.Join(appsDir, appName)
	if !config.IsWithinDirectory(cleanPath, appDir) {
		return ReadFileResult{}, fmt.Errorf("file path attempts to access file outside app directory")
	}

	// Get file info
	info, err := os.Lstat(cleanPath) // Use Lstat to detect symlinks
	if err != nil {
		return ReadFileResult{}, err
	}

	// Check for symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		return ReadFileResult{}, ErrSymlink
	}

	// Check file size
	if info.Size() > constants.MaxFileSize {
		return ReadFileResult{}, ErrFileTooLarge
	}

	// Read file content and check if binary in one operation
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return ReadFileResult{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if content is binary using the actual file content
	if len(content) > 0 {
		// Use first 512 bytes (or less if file is smaller) for content type detection
		sampleSize := len(content)
		if sampleSize > 512 {
			sampleSize = 512
		}
		contentType := http.DetectContentType(content[:sampleSize])
		if !strings.HasPrefix(contentType, "text/") {
			return ReadFileResult{}, ErrBinaryFile
		}
	}

	return ReadFileResult{
		AppName:  appName,
		FilePath: filePath,
		Content:  string(content),
	}, nil
}

// CLI
func ReadFileCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printReadFileHelp()
			return nil
		}
	}

	var appName, filePath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--app-name":
			if i+1 < len(args) {
				appName = args[i+1]
				i++
			} else {
				return errors.New("--app-name requires a value")
			}
		case "--file-path":
			if i+1 < len(args) {
				filePath = args[i+1]
				i++
			} else {
				return errors.New("--file-path requires a value")
			}
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool read_file --help' for usage", args[i])
			}
		}
	}

	if appName == "" {
		return errors.New("--app-name is required")
	}
	if filePath == "" {
		return errors.New("--file-path is required")
	}

	result, err := ReadFile(appName, filePath)
	if err != nil {
		return err
	}

	fmt.Printf("App: %s\nFile: %s\n\nContent:\n%s\n", result.AppName, result.FilePath, result.Content)
	return nil
}

func printReadFileHelp() {
	fmt.Println("Usage: layered-code tool read_file [options]")
	fmt.Println()
	fmt.Println("Read the contents of a file within an application directory")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>    Name of the app directory")
	fmt.Println("  --file-path <path>   Path to the file relative to the app directory")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Symlinks are not followed")
	fmt.Println("  - Binary files are not supported")
	fmt.Printf("  - Maximum file size is %s\n", constants.MaxFileSizeInWords)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Read a source file")
	fmt.Println("  layered-code tool read_file --app-name myapp --file-path src/main.go")
	fmt.Println()
	fmt.Println("  # Read a configuration file")
	fmt.Println("  layered-code tool read_file --app-name myapp --file-path config/settings.json")
}

// MCP
func ReadFileMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName  string `json:"app_name"`
		FilePath string `json:"file_path"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := ReadFile(args.AppName, args.FilePath)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}
