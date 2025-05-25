package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/mark3labs/mcp-go/mcp"
)

type ListFilesResult struct {
	AppName string      `json:"app_name"`
	AppPath string      `json:"app_path"`
	Files   []FileEntry `json:"files"`
}

type FileEntry struct {
	Path         string     `json:"path"`
	Name         string     `json:"name"`
	IsDirectory  bool       `json:"is_directory"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Size         *string    `json:"size,omitempty"`
	ChildCount   *int       `json:"child_count,omitempty"`
}

var sizeCache = make(map[string]int64)
var sizeCacheMutex sync.RWMutex

func ListFiles(appName string, pattern *string, includeLastModified, includeSize, includeChildCount bool) (ListFilesResult, error) {
	if appName == "" {
		return ListFilesResult{}, errors.New("app_name is required")
	}

	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return ListFilesResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}
	appPath := filepath.Join(appsDir, appName)

	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return ListFilesResult{}, fmt.Errorf("app '%s' not found in apps directory", appName)
	}

	// Validate pattern if provided
	if pattern != nil && *pattern != "" {
		if strings.Contains(*pattern, "..") {
			return ListFilesResult{}, errors.New("invalid pattern: directory traversal is not allowed")
		}
	}

	var entries []FileEntry

	err = walkWithDepth(appPath, appPath, func(path string, info os.FileInfo, currentDepth int) error {
		// Check max depth
		if currentDepth > constants.MaxDirectoryDepth {
			return filepath.SkipDir
		}

		// Skip hidden files/folders (starting with .)
		if strings.HasPrefix(info.Name(), ".") && path != appPath {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		relPath, err := filepath.Rel(appPath, path)
		if err != nil {
			return err
		}

		entry := FileEntry{
			Path:        relPath,
			Name:        info.Name(),
			IsDirectory: info.IsDir(),
		}

		if includeLastModified {
			modTime := info.ModTime()
			entry.LastModified = &modTime
		}

		if includeSize {
			size := info.Size()
			if info.IsDir() {
				size = getCachedDirSize(path)
			}
			sizeStr := formatSize(size)
			entry.Size = &sizeStr
		}

		if includeChildCount {
			count := 0
			if info.IsDir() {
				count = getChildCount(path)
			}
			entry.ChildCount = &count
		}

		// Apply glob pattern filter if specified
		if pattern != nil && *pattern != "" {
			matched := false

			// Handle ** pattern for recursive matching
			if strings.Contains(*pattern, "**") {
				// Convert ** pattern to work with filepath.Match
				// e.g., "**/*.js" becomes a check for suffix ".js"
				parts := strings.Split(*pattern, "**/")
				if len(parts) == 2 && parts[0] == "" {
					// Pattern like "**/*.js"
					subPattern := parts[1]
					matched, _ = filepath.Match(subPattern, filepath.Base(relPath))
				}
			} else {
				// Standard glob pattern matching
				matched, err = filepath.Match(*pattern, relPath)
				if err != nil {
					return fmt.Errorf("invalid glob pattern: %w", err)
				}
				if !matched {
					// Also try matching just the filename
					matched, _ = filepath.Match(*pattern, info.Name())
				}
			}

			if !matched {
				return nil
			}
		}

		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		return ListFilesResult{}, err
	}

	return ListFilesResult{
		AppName: appName,
		AppPath: appPath,
		Files:   entries,
	}, nil
}

func walkWithDepth(root, basePath string, fn func(path string, info os.FileInfo, depth int) error) error {
	depth := 0
	if basePath != "" {
		relPath, _ := filepath.Rel(basePath, root)
		depth = strings.Count(relPath, string(os.PathSeparator))
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(basePath, path)
		currentDepth := depth + strings.Count(relPath, string(os.PathSeparator))

		return fn(path, info, currentDepth)
	})
}

func getCachedDirSize(path string) int64 {
	sizeCacheMutex.RLock()
	if size, ok := sizeCache[path]; ok {
		sizeCacheMutex.RUnlock()
		return size
	}
	sizeCacheMutex.RUnlock()

	size, _ := calculateDirSize(path)

	sizeCacheMutex.Lock()
	sizeCache[path] = size
	sizeCacheMutex.Unlock()

	return size
}

func calculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip hidden files/folders
		if strings.HasPrefix(info.Name(), ".") && filePath != path {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func getChildCount(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	// Count non-hidden entries
	count := 0
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			count++
		}
	}
	return count
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func ListFilesCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printListFilesHelp()
			return nil
		}
	}

	var appName string
	var pattern *string
	var includeLastModified, includeSize, includeChildCount bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--app-name":
			if i+1 < len(args) {
				appName = args[i+1]
				i++
			} else {
				return errors.New("--app-name requires a value")
			}
		case "--pattern":
			if i+1 < len(args) {
				pattern = &args[i+1]
				i++
			} else {
				return errors.New("--pattern requires a value")
			}
		case "--include-last-modified":
			includeLastModified = true
		case "--include-size":
			includeSize = true
		case "--include-child-count":
			includeChildCount = true
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool list_files --help' for usage", args[i])
			}
		}
	}

	if appName == "" {
		return errors.New("--app-name is required")
	}

	result, err := ListFiles(appName, pattern, includeLastModified, includeSize, includeChildCount)
	if err != nil {
		return err
	}

	// Output in human-readable format
	fmt.Printf("App: %s\nPath: %s\n\n", result.AppName, result.AppPath)

	for _, file := range result.Files {
		// Indent based on path depth
		depth := strings.Count(file.Path, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)

		if file.IsDirectory {
			fmt.Printf("%s%s/ ", indent, file.Name)
		} else {
			fmt.Printf("%s%s ", indent, file.Name)
		}

		// Add optional metadata
		var metadata []string
		if file.Size != nil {
			metadata = append(metadata, *file.Size)
		}
		if file.LastModified != nil {
			metadata = append(metadata, file.LastModified.Format("2006-01-02 15:04:05"))
		}
		if file.ChildCount != nil && file.IsDirectory {
			metadata = append(metadata, fmt.Sprintf("%d items", *file.ChildCount))
		}

		if len(metadata) > 0 {
			fmt.Printf("(%s)", strings.Join(metadata, ", "))
		}
		fmt.Println()
	}

	return nil
}

func printListFilesHelp() {
	fmt.Println("Usage: layered-code tool list_files [options]")
	fmt.Println()
	fmt.Println("List files and directories within an application")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>          Name of the app directory to analyze")
	fmt.Println()
	fmt.Println("Optional options:")
	fmt.Println("  --pattern <glob>           Filter files using glob pattern (e.g. '*.txt', 'src/*.js')")
	fmt.Println("  --include-last-modified    Include last modification timestamps")
	fmt.Println("  --include-size             Include file/directory sizes in human-readable format")
	fmt.Println("  --include-child-count      Include count of immediate children for directories")
	fmt.Println("  --help, -h                 Show this help message")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Hidden files/folders (starting with '.') and symlinks are automatically skipped")
	fmt.Printf("  - Maximum directory depth is limited to %d levels for safety\n", constants.MaxDirectoryDepth)
	fmt.Println("  - Directory sizes are cached for performance")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List basic files")
	fmt.Println("  layered-code tool list_files --app-name myapp")
	fmt.Println()
	fmt.Println("  # List files with all metadata")
	fmt.Println("  layered-code tool list_files --app-name myapp --include-size --include-last-modified --include-child-count")
	fmt.Println()
	fmt.Println("  # List files matching a pattern")
	fmt.Println("  layered-code tool list_files --app-name myapp --pattern '*.js'")
	fmt.Println()
	fmt.Println("  # List files in specific subdirectory")
	fmt.Println("  layered-code tool list_files --app-name myapp --pattern 'src/*.go'")
}

func ListFilesMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName             string  `json:"app_name"`
		Pattern             *string `json:"pattern"`
		IncludeLastModified bool    `json:"include_last_modified"`
		IncludeSize         bool    `json:"include_size"`
		IncludeChildCount   bool    `json:"include_child_count"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	result, err := ListFiles(args.AppName, args.Pattern, args.IncludeLastModified, args.IncludeSize, args.IncludeChildCount)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}
