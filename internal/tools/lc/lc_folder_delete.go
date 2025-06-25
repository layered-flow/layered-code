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

// LcFolderDeleteParams represents the parameters for deleting a folder
type LcFolderDeleteParams struct {
	AppName    string `json:"app_name"`
	FolderPath string `json:"folder_path"`
}

// LcFolderDeleteResult represents the result of a folder delete operation
type LcFolderDeleteResult struct {
	AppName    string `json:"app_name"`
	FolderPath string `json:"folder_path"`
	Deleted    bool   `json:"deleted"`
}

// LcFolderDelete deletes a folder and all its contents within an app directory
func LcFolderDelete(params LcFolderDeleteParams) (LcFolderDeleteResult, error) {
	if params.AppName == "" {
		return LcFolderDeleteResult{}, errors.New("app_name is required")
	}
	if params.FolderPath == "" {
		return LcFolderDeleteResult{}, errors.New("folder_path is required")
	}

	// Validate path doesn't contain directory traversal
	if strings.Contains(params.FolderPath, "..") {
		return LcFolderDeleteResult{}, errors.New("directory traversal is not allowed")
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcFolderDeleteResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct full path
	appPath := filepath.Join(appsDir, params.AppName)
	fullPath := filepath.Join(appPath, params.FolderPath)

	// Ensure path is within the app directory
	cleanPath := filepath.Clean(fullPath)
	if !config.IsWithinDirectory(cleanPath, appPath) {
		return LcFolderDeleteResult{}, errors.New("path must be within the app directory")
	}

	// Don't allow deleting the app root directory itself
	if cleanPath == appPath {
		return LcFolderDeleteResult{}, errors.New("cannot delete the app root directory")
	}

	// Check if folder exists
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return LcFolderDeleteResult{}, fmt.Errorf("folder not found: %s", params.FolderPath)
		}
		return LcFolderDeleteResult{}, fmt.Errorf("error accessing folder: %w", err)
	}

	// Ensure it's a directory
	if !fileInfo.IsDir() {
		return LcFolderDeleteResult{}, errors.New("path is not a directory, use lc_file_delete for files")
	}

	// Delete the folder and all its contents
	if err := os.RemoveAll(cleanPath); err != nil {
		return LcFolderDeleteResult{}, fmt.Errorf("failed to delete folder: %w", err)
	}

	// Send notification
	notificationPath := filepath.Join(params.AppName, params.FolderPath)
	notifications.NotifyFileChange(notificationPath, "deleted")

	return LcFolderDeleteResult{
		AppName:    params.AppName,
		FolderPath: params.FolderPath,
		Deleted:    true,
	}, nil
}

// CLI
func LcFolderDeleteCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printLcFolderDeleteHelp()
			return nil
		}
	}

	var params LcFolderDeleteParams
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
		case "--folder-path":
			if i+1 < len(args) {
				params.FolderPath = args[i+1]
				i++
			} else {
				return errors.New("--folder-path requires a value")
			}
		case "--force", "-f":
			force = true
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool lc_folder_delete --help' for usage", args[i])
			}
		}
	}

	// Validate required parameters
	if params.AppName == "" || params.FolderPath == "" {
		return errors.New("both --app-name and --folder-path are required")
	}

	// Confirm deletion if not forced
	if !force {
		fmt.Printf("Are you sure you want to delete the folder '%s/%s' and all its contents? This action cannot be undone. [y/N]: ", params.AppName, params.FolderPath)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	result, err := LcFolderDelete(params)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted folder: %s/%s\n", result.AppName, result.FolderPath)
	return nil
}

func printLcFolderDeleteHelp() {
	fmt.Println("Usage: layered-code tool lc_folder_delete [options]")
	fmt.Println()
	fmt.Println("Delete a folder and all its contents within an application directory")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>       Name of the app directory")
	fmt.Println("  --folder-path <path>    Path to the folder relative to the app directory")
	fmt.Println()
	fmt.Println("Optional options:")
	fmt.Println("  --force, -f             Skip confirmation prompt")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - This will delete the folder and ALL its contents recursively")
	fmt.Println("  - This action cannot be undone")
	fmt.Println("  - Cannot delete the app root directory itself")
	fmt.Println("  - Without --force, you will be prompted to confirm")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Delete a folder with confirmation")
	fmt.Println("  layered-code tool lc_folder_delete --app-name myapp --folder-path old-feature")
	fmt.Println()
	fmt.Println("  # Delete a nested folder without confirmation")
	fmt.Println("  layered-code tool lc_folder_delete --app-name myapp --folder-path src/deprecated --force")
}

// MCP
func LcFolderDeleteMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LcFolderDeleteParams

	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	result, err := LcFolderDelete(params)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}

