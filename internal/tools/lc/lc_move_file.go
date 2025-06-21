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

// LcMoveFileParams represents the parameters for moving/renaming a file
type LcMoveFileParams struct {
	AppName     string `json:"app_name"`
	SourcePath  string `json:"source_path"`
	DestPath    string `json:"dest_path"`
}

// LcMoveFileResult represents the result of a move/rename operation
type LcMoveFileResult struct {
	AppName     string `json:"app_name"`
	SourcePath  string `json:"source_path"`
	DestPath    string `json:"dest_path"`
	IsRename    bool   `json:"is_rename"`
}

// LcMoveFile moves or renames a file within an app directory
func LcMoveFile(params LcMoveFileParams) (LcMoveFileResult, error) {
	if params.AppName == "" {
		return LcMoveFileResult{}, errors.New("app_name is required")
	}
	if params.SourcePath == "" {
		return LcMoveFileResult{}, errors.New("source_path is required")
	}
	if params.DestPath == "" {
		return LcMoveFileResult{}, errors.New("dest_path is required")
	}

	// Validate paths don't contain directory traversal
	if strings.Contains(params.SourcePath, "..") || strings.Contains(params.DestPath, "..") {
		return LcMoveFileResult{}, errors.New("directory traversal is not allowed")
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return LcMoveFileResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct full paths
	appPath := filepath.Join(appsDir, params.AppName)
	sourcePath := filepath.Join(appPath, params.SourcePath)
	destPath := filepath.Join(appPath, params.DestPath)

	// Ensure paths are within the app directory
	cleanSourcePath := filepath.Clean(sourcePath)
	cleanDestPath := filepath.Clean(destPath)
	if !strings.HasPrefix(cleanSourcePath, appPath) || !strings.HasPrefix(cleanDestPath, appPath) {
		return LcMoveFileResult{}, errors.New("paths must be within the app directory")
	}

	// Check if source file exists
	sourceInfo, err := os.Stat(cleanSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return LcMoveFileResult{}, fmt.Errorf("source file not found: %s", params.SourcePath)
		}
		return LcMoveFileResult{}, fmt.Errorf("error accessing source file: %w", err)
	}

	// Don't allow moving directories (for now)
	if sourceInfo.IsDir() {
		return LcMoveFileResult{}, errors.New("moving directories is not supported")
	}

	// Check if destination exists
	if _, err := os.Stat(cleanDestPath); err == nil {
		return LcMoveFileResult{}, fmt.Errorf("destination already exists: %s", params.DestPath)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(cleanDestPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return LcMoveFileResult{}, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Perform the move/rename
	if err := os.Rename(cleanSourcePath, cleanDestPath); err != nil {
		return LcMoveFileResult{}, fmt.Errorf("failed to move file: %w", err)
	}

	// Determine if this was a rename (same directory) or move
	isRename := filepath.Dir(params.SourcePath) == filepath.Dir(params.DestPath)

	// Send notifications
	sourceNotificationPath := filepath.Join(params.AppName, params.SourcePath)
	destNotificationPath := filepath.Join(params.AppName, params.DestPath)
	notifications.NotifyFileChange(sourceNotificationPath, "deleted")
	notifications.NotifyFileChange(destNotificationPath, "created")

	return LcMoveFileResult{
		AppName:    params.AppName,
		SourcePath: params.SourcePath,
		DestPath:   params.DestPath,
		IsRename:   isRename,
	}, nil
}

// CLI
func LcMoveFileCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printLcMoveFileHelp()
			return nil
		}
	}

	var params LcMoveFileParams

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
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool lc_move_file --help' for usage", args[i])
			}
		}
	}

	result, err := LcMoveFile(params)
	if err != nil {
		return err
	}

	if result.IsRename {
		fmt.Printf("Renamed: %s/%s -> %s\n", result.AppName, result.SourcePath, result.DestPath)
	} else {
		fmt.Printf("Moved: %s/%s -> %s/%s\n", result.AppName, result.SourcePath, result.AppName, result.DestPath)
	}
	return nil
}

func printLcMoveFileHelp() {
	fmt.Println("Usage: layered-code tool lc_move_file [options]")
	fmt.Println()
	fmt.Println("Move or rename a file within an application directory")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>    Name of the app directory")
	fmt.Println("  --source <path>      Source file path relative to the app directory")
	fmt.Println("  --dest <path>        Destination file path relative to the app directory")
	fmt.Println()
	fmt.Println("Aliases:")
	fmt.Println("  --from               Alias for --source")
	fmt.Println("  --to                 Alias for --dest")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Moving directories is not supported")
	fmt.Println("  - Destination must not already exist")
	fmt.Println("  - Parent directories will be created if needed")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Rename a file in the same directory")
	fmt.Println("  layered-code tool lc_move_file --app-name myapp --source old.txt --dest new.txt")
	fmt.Println()
	fmt.Println("  # Move a file to a different directory")
	fmt.Println("  layered-code tool lc_move_file --app-name myapp --from src/old.js --to archive/old.js")
}

// MCP
func LcMoveFileMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params LcMoveFileParams

	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	result, err := LcMoveFile(params)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}