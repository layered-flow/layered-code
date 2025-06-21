package lc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/layered-flow/layered-code/internal/notifications"
	"github.com/mark3labs/mcp-go/mcp"
)

// LcCopyFileParams represents the parameters for copying a file
type LcCopyFileParams struct {
	AppName     string `json:"app_name"`
	SourcePath  string `json:"source_path"`
	DestPath    string `json:"dest_path"`
	Overwrite   bool   `json:"overwrite"`
}

// LcCopyFileResult represents the result of a copy operation
type LcCopyFileResult struct {
	AppName     string `json:"app_name"`
	SourcePath  string `json:"source_path"`
	DestPath    string `json:"dest_path"`
	BytesCopied int64  `json:"bytes_copied"`
}

// LcCopyFile copies a file within an app directory
func LcCopyFile(params LcCopyFileParams) (LcCopyFileResult, error) {
	if params.AppName == "" {
		return LcCopyFileResult{}, errors.New("app_name is required")
	}
	if params.SourcePath == "" {
		return LcCopyFileResult{}, errors.New("source_path is required")
	}
	if params.DestPath == "" {
		return LcCopyFileResult{}, errors.New("dest_path is required")
	}

	// Validate paths don't contain directory traversal
	if strings.Contains(params.SourcePath, "..") || strings.Contains(params.DestPath, "..") {
		return LcCopyFileResult{}, errors.New("directory traversal is not allowed")
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcCopyFileResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct full paths
	appPath := filepath.Join(appsDir, params.AppName)
	sourcePath := filepath.Join(appPath, params.SourcePath)
	destPath := filepath.Join(appPath, params.DestPath)

	// Ensure paths are within the app directory
	cleanSourcePath := filepath.Clean(sourcePath)
	cleanDestPath := filepath.Clean(destPath)
	if !config.IsWithinDirectory(cleanSourcePath, appPath) || !config.IsWithinDirectory(cleanDestPath, appPath) {
		return LcCopyFileResult{}, errors.New("paths must be within the app directory")
	}

	// Check if source file exists
	sourceInfo, err := os.Stat(cleanSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return LcCopyFileResult{}, fmt.Errorf("source file not found: %s", params.SourcePath)
		}
		return LcCopyFileResult{}, fmt.Errorf("error accessing source file: %w", err)
	}

	// Don't allow copying directories
	if sourceInfo.IsDir() {
		return LcCopyFileResult{}, errors.New("copying directories is not supported")
	}

	// Check file size limit
	if sourceInfo.Size() > constants.MaxFileSize {
		return LcCopyFileResult{}, fmt.Errorf("file exceeds maximum size of %s", constants.MaxFileSizeInWords)
	}

	// Prevent copying file to itself
	if cleanSourcePath == cleanDestPath {
		return LcCopyFileResult{}, errors.New("source and destination are the same")
	}

	// Check if destination exists
	if _, err := os.Stat(cleanDestPath); err == nil && !params.Overwrite {
		return LcCopyFileResult{}, fmt.Errorf("destination already exists: %s (use overwrite option to replace)", params.DestPath)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(cleanDestPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return LcCopyFileResult{}, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	sourceFile, err := os.Open(cleanSourcePath)
	if err != nil {
		return LcCopyFileResult{}, fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(cleanDestPath)
	if err != nil {
		return LcCopyFileResult{}, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the file contents
	bytesCopied, err := io.Copy(destFile, sourceFile)
	if err != nil {
		// Try to clean up on error
		os.Remove(cleanDestPath)
		return LcCopyFileResult{}, fmt.Errorf("failed to copy file: %w", err)
	}

	// Copy file permissions
	if err := os.Chmod(cleanDestPath, sourceInfo.Mode()); err != nil {
		// Non-fatal, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to copy file permissions: %v\n", err)
	}

	// Send notification
	notificationPath := filepath.Join(params.AppName, params.DestPath)
	notifications.NotifyFileChange(notificationPath, "created")

	return LcCopyFileResult{
		AppName:     params.AppName,
		SourcePath:  params.SourcePath,
		DestPath:    params.DestPath,
		BytesCopied: bytesCopied,
	}, nil
}

// CLI
func LcCopyFileCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printLcCopyFileHelp()
			return nil
		}
	}

	var params LcCopyFileParams

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--app-name":
			if i+1 < len(args) {
				params.AppName = args[i+1]
				i++
			} else {
				return errors.New("--app-name requires a value")
			}
		case "--source", "--from":
			if i+1 < len(args) {
				params.SourcePath = args[i+1]
				i++
			} else {
				return errors.New("--source requires a value")
			}
		case "--dest", "--to":
			if i+1 < len(args) {
				params.DestPath = args[i+1]
				i++
			} else {
				return errors.New("--dest requires a value")
			}
		case "--overwrite", "-f":
			params.Overwrite = true
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool lc_copy_file --help' for usage", args[i])
			}
		}
	}

	result, err := LcCopyFile(params)
	if err != nil {
		return err
	}

	fmt.Printf("Copied: %s/%s -> %s/%s\n", result.AppName, result.SourcePath, result.AppName, result.DestPath)
	fmt.Printf("Bytes copied: %d\n", result.BytesCopied)
	return nil
}

func printLcCopyFileHelp() {
	fmt.Println("Usage: layered-code tool lc_copy_file [options]")
	fmt.Println()
	fmt.Println("Copy a file within an application directory")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>    Name of the app directory")
	fmt.Println("  --source <path>      Source file path relative to the app directory")
	fmt.Println("  --dest <path>        Destination file path relative to the app directory")
	fmt.Println()
	fmt.Println("Optional options:")
	fmt.Println("  --overwrite, -f      Overwrite destination if it exists")
	fmt.Println()
	fmt.Println("Aliases:")
	fmt.Println("  --from               Alias for --source")
	fmt.Println("  --to                 Alias for --dest")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Copying directories is not supported")
	fmt.Println("  - File permissions are preserved when possible")
	fmt.Printf("  - Maximum file size is %s\n", constants.MaxFileSizeInWords)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Copy a file to a new location")
	fmt.Println("  layered-code tool lc_copy_file --app-name myapp --source config.json --dest config.backup.json")
	fmt.Println()
	fmt.Println("  # Copy and overwrite existing file")
	fmt.Println("  layered-code tool lc_copy_file --app-name myapp --from template.html --to index.html --overwrite")
}

// MCP
func LcCopyFileMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LcCopyFileParams

	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	result, err := LcCopyFile(params)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}