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

// EditFileParams represents the parameters for editing a file
type EditFileParams struct {
	AppName      string `json:"app_name"`
	FilePath     string `json:"file_path"`
	OldString    string `json:"old_string"`
	NewString    string `json:"new_string"`
	Occurrences  int    `json:"occurrences"`  // Number of occurrences to replace (0 = all)
}

// EditFileResult represents the result of editing a file
type EditFileResult struct {
	AppName        string     `json:"app_name"`
	FilePath       string     `json:"file_path"`
	Replacements   int        `json:"replacements"`
	LastModified   *time.Time `json:"last_modified,omitempty"`
}

// EditFile performs find-and-replace operations on a file within an app directory
func EditFile(params EditFileParams) (EditFileResult, error) {
	if params.AppName == "" {
		return EditFileResult{}, errors.New("app_name is required")
	}
	if params.FilePath == "" {
		return EditFileResult{}, errors.New("file_path is required")
	}
	if params.OldString == "" {
		return EditFileResult{}, errors.New("old_string is required")
	}
	if params.Occurrences < 0 {
		return EditFileResult{}, errors.New("occurrences must be non-negative")
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return EditFileResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct and validate the full file path
	fullPath := filepath.Join(appsDir, params.AppName, params.FilePath)
	cleanPath := filepath.Clean(fullPath)

	// Ensure the file is within the app directory
	appDir := filepath.Join(appsDir, params.AppName)
	if !config.IsWithinDirectory(cleanPath, appDir) {
		return EditFileResult{}, fmt.Errorf("file path attempts to access file outside app directory")
	}

	// Read the file
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return EditFileResult{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Check file size
	if len(content) > int(constants.MaxFileSize) {
		return EditFileResult{}, fmt.Errorf("file exceeds maximum size of %s", constants.MaxFileSizeInWords)
	}

	// Convert to string for editing
	fileContent := string(content)

	// Count occurrences
	totalOccurrences := strings.Count(fileContent, params.OldString)
	if totalOccurrences == 0 {
		return EditFileResult{}, fmt.Errorf("old_string not found in file")
	}

	// Perform replacements
	replacements := 0
	if params.Occurrences == 0 {
		// Replace all occurrences
		fileContent = strings.ReplaceAll(fileContent, params.OldString, params.NewString)
		replacements = totalOccurrences
	} else {
		// Replace specific number of occurrences
		maxReplacements := params.Occurrences
		if maxReplacements > totalOccurrences {
			maxReplacements = totalOccurrences
		}

		result := strings.Builder{}
		remaining := fileContent
		for i := 0; i < maxReplacements; i++ {
			index := strings.Index(remaining, params.OldString)
			if index == -1 {
				break
			}
			result.WriteString(remaining[:index])
			result.WriteString(params.NewString)
			remaining = remaining[index+len(params.OldString):]
			replacements++
		}
		result.WriteString(remaining)
		fileContent = result.String()
	}

	// Write the modified content back
	if err := os.WriteFile(cleanPath, []byte(fileContent), 0644); err != nil {
		return EditFileResult{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Send WebSocket notification
	notificationPath := filepath.Join(params.AppName, params.FilePath)
	notifications.NotifyFileChange(notificationPath, "edit")

	// Get file info for the result
	info, err := os.Stat(cleanPath)
	if err != nil {
		return EditFileResult{}, fmt.Errorf("failed to stat edited file: %w", err)
	}

	modTime := info.ModTime()
	return EditFileResult{
		AppName:      params.AppName,
		FilePath:     params.FilePath,
		Replacements: replacements,
		LastModified: &modTime,
	}, nil
}

// CLI
func EditFileCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printEditFileHelp()
			return nil
		}
	}

	var params EditFileParams

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
		case "--old-string":
			if i+1 < len(args) {
				params.OldString = args[i+1]
				i++
			} else {
				return errors.New("--old-string requires a value")
			}
		case "--new-string":
			if i+1 < len(args) {
				params.NewString = args[i+1]
				i++
			} else {
				return errors.New("--new-string requires a value")
			}
		case "--occurrences":
			if i+1 < len(args) {
				var err error
				fmt.Sscanf(args[i+1], "%d", &params.Occurrences)
				if err != nil {
					return fmt.Errorf("--occurrences must be a number: %v", err)
				}
				i++
			} else {
				return errors.New("--occurrences requires a value")
			}
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool edit_file --help' for usage", args[i])
			}
		}
	}

	if params.AppName == "" {
		return errors.New("--app-name is required")
	}
	if params.FilePath == "" {
		return errors.New("--file-path is required")
	}
	if params.OldString == "" {
		return errors.New("--old-string is required")
	}
	// Note: new-string can be empty (for deletion)

	result, err := EditFile(params)
	if err != nil {
		return err
	}

	fmt.Printf("Edited file: %s/%s\n", result.AppName, result.FilePath)
	fmt.Printf("Replacements made: %d\n", result.Replacements)
	return nil
}

func printEditFileHelp() {
	fmt.Println("Usage: layered-code tool edit_file [options]")
	fmt.Println()
	fmt.Println("Edit a file by performing find-and-replace operations")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>      Name of the app directory")
	fmt.Println("  --file-path <path>     Path to the file relative to the app directory")
	fmt.Println("  --old-string <text>    Text to find and replace")
	fmt.Println("  --new-string <text>    Text to replace with (can be empty for deletion)")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  --occurrences <num>    Number of occurrences to replace (0 = all, default: 0)")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - The file must be a text file")
	fmt.Printf("  - Maximum file size is %s\n", constants.MaxFileSizeInWords)
	fmt.Println("  - Use --new-string \"\" to delete text")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Replace all occurrences")
	fmt.Println("  layered-code tool edit_file --app-name myapp --file-path config.json \\")
	fmt.Println("    --old-string 'localhost' --new-string '127.0.0.1'")
	fmt.Println()
	fmt.Println("  # Replace first 2 occurrences")
	fmt.Println("  layered-code tool edit_file --app-name myapp --file-path src/main.go \\")
	fmt.Println("    --old-string 'fmt.Println' --new-string 'log.Println' --occurrences 2")
	fmt.Println()
	fmt.Println("  # Delete text")
	fmt.Println("  layered-code tool edit_file --app-name myapp --file-path README.md \\")
	fmt.Println("    --old-string 'TODO: ' --new-string ''")
}

// MCP
func EditFileMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params EditFileParams

	if err := request.BindArguments(&params); err != nil {
		return nil, err
	}

	result, err := EditFile(params)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}