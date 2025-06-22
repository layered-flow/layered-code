package lc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/notifications"
	"github.com/mark3labs/mcp-go/mcp"
)

// LcFileDeleteParams represents the parameters for deleting a file
type LcFileDeleteParams struct {
	AppName  string `json:"app_name"`
	FilePath string `json:"file_path"`
}

// LcFileDeleteResult represents the result of a delete operation
type LcFileDeleteResult struct {
	AppName  string `json:"app_name"`
	FilePath string `json:"file_path"`
	Deleted  bool   `json:"deleted"`
}

// LcFileDelete deletes a file within an app directory
func LcFileDelete(params LcFileDeleteParams) (LcFileDeleteResult, error) {
	if params.AppName == "" {
		return LcFileDeleteResult{}, errors.New("app_name is required")
	}
	if params.FilePath == "" {
		return LcFileDeleteResult{}, errors.New("file_path is required")
	}

	// Validate path doesn't contain directory traversal
	if strings.Contains(params.FilePath, "..") {
		return LcFileDeleteResult{}, errors.New("directory traversal is not allowed")
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcFileDeleteResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct full path
	appPath := filepath.Join(appsDir, params.AppName)
	fullPath := filepath.Join(appPath, params.FilePath)

	// Ensure path is within the app directory
	cleanPath := filepath.Clean(fullPath)
	if !config.IsWithinDirectory(cleanPath, appPath) {
		return LcFileDeleteResult{}, errors.New("path must be within the app directory")
	}

	// Check if file exists
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return LcFileDeleteResult{}, fmt.Errorf("file not found: %s", params.FilePath)
		}
		return LcFileDeleteResult{}, fmt.Errorf("error accessing file: %w", err)
	}

	// Don't allow deleting directories
	if fileInfo.IsDir() {
		return LcFileDeleteResult{}, errors.New("deleting directories is not supported, only files can be deleted")
	}

	// Delete the file
	if err := os.Remove(cleanPath); err != nil {
		return LcFileDeleteResult{}, fmt.Errorf("failed to delete file: %w", err)
	}

	// Send notification
	notificationPath := filepath.Join(params.AppName, params.FilePath)
	notifications.NotifyFileChange(notificationPath, "deleted")

	return LcFileDeleteResult{
		AppName:  params.AppName,
		FilePath: params.FilePath,
		Deleted:  true,
	}, nil
}

// CLI
func LcFileDeleteCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printLcFileDeleteHelp()
			return nil
		}
	}

	var params LcFileDeleteParams
	var force bool

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
		case "--force", "-f":
			force = true
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool lc_file_delete --help' for usage", args[i])
			}
		}
	}

	// Validate required parameters
	if params.AppName == "" || params.FilePath == "" {
		return errors.New("both --app-name and --file-path are required")
	}

	// Confirm deletion if not forced
	if !force {
		fmt.Printf("Are you sure you want to delete '%s/%s'? This action cannot be undone. [y/N]: ", params.AppName, params.FilePath)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	result, err := LcFileDelete(params)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted: %s/%s\n", result.AppName, result.FilePath)
	return nil
}

func printLcFileDeleteHelp() {
	fmt.Println("Usage: layered-code tool lc_file_delete [options]")
	fmt.Println()
	fmt.Println("Delete a file within an application directory")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>    Name of the app directory")
	fmt.Println("  --file-path <path>   Path to the file relative to the app directory")
	fmt.Println()
	fmt.Println("Optional options:")
	fmt.Println("  --force, -f          Skip confirmation prompt")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Only files can be deleted, not directories")
	fmt.Println("  - This action cannot be undone")
	fmt.Println("  - Without --force, you will be prompted to confirm")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Delete a file with confirmation")
	fmt.Println("  layered-code tool lc_file_delete --app-name myapp --file-path old-file.txt")
	fmt.Println()
	fmt.Println("  # Delete a file without confirmation")
	fmt.Println("  layered-code tool lc_file_delete --app-name myapp --file-path temp.log --force")
}

// MCP
func LcFileDeleteMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LcFileDeleteParams

	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	result, err := LcFileDelete(params)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}