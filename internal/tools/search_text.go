package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/layered-flow/layered-code/internal/config"
	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/mark3labs/mcp-go/mcp"
)

// SearchTextResult represents the result of searching for text in files
type SearchTextResult struct {
	AppName string        `json:"app_name"`
	Pattern string        `json:"pattern"`
	Matches []SearchMatch `json:"matches"`
	Total   int           `json:"total_matches"`
}

// SearchMatch represents a single search match
type SearchMatch struct {
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	LineText   string `json:"line_text"`
	Match      string `json:"match"`
}

// SearchTextOptions configures the search behavior
type SearchTextOptions struct {
	CaseSensitive bool
	WholeWord     bool
	FilePattern   string
	MaxResults    int
	IncludeHidden bool
}

// SearchText searches for a pattern in files within an app directory using ripgrep
func SearchText(appName, pattern string, options SearchTextOptions) (SearchTextResult, error) {
	if appName == "" {
		return SearchTextResult{}, errors.New("app_name is required")
	}
	if pattern == "" {
		return SearchTextResult{}, errors.New("pattern is required")
	}

	// Get and validate the apps directory
	appsDir, err := config.EnsureAppsDirectory()
	if err != nil {
		return SearchTextResult{}, fmt.Errorf("failed to ensure apps directory: %w", err)
	}

	// Construct and validate the app directory path
	appDir := filepath.Join(appsDir, appName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return SearchTextResult{}, fmt.Errorf("app '%s' not found in apps directory", appName)
	}
	
	// Use build directory if it exists, otherwise fall back to app directory
	searchDir := appDir
	outputDir := filepath.Join(appDir, constants.OutputDirectoryName)
	if _, err := os.Stat(outputDir); err == nil {
		searchDir = outputDir
	}

	// Get ripgrep binary path
	rgPath, err := getRipgrepPath()
	if err != nil {
		return SearchTextResult{}, err
	}

	// Build ripgrep command arguments
	args := []string{
		"--json",          // JSON output for easy parsing
		"--line-number",   // Include line numbers
		"--with-filename", // Include filenames
		"--no-heading",    // Don't group matches by file
	}

	// Add options
	if !options.CaseSensitive {
		args = append(args, "--ignore-case")
	}
	if options.WholeWord {
		args = append(args, "--word-regexp")
	}
	if options.FilePattern != "" {
		args = append(args, "--glob", options.FilePattern)
	}
	if options.MaxResults > 0 {
		args = append(args, "--max-count", fmt.Sprintf("%d", options.MaxResults))
	}
	if options.IncludeHidden {
		args = append(args, "--hidden")
	}

	// Add pattern and path
	args = append(args, "--", pattern, searchDir)

	// Execute ripgrep
	cmd := exec.Command(rgPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		// Exit code 1 means no matches found, which is not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return SearchTextResult{
				AppName: appName,
				Pattern: pattern,
				Matches: []SearchMatch{},
				Total:   0,
			}, nil
		}
		return SearchTextResult{}, fmt.Errorf("ripgrep failed: %w (stderr: %s)", err, stderr.String())
	}

	// Parse ripgrep JSON output
	matches, err := parseRipgrepOutput(stdout.String(), appDir)
	if err != nil {
		return SearchTextResult{}, fmt.Errorf("failed to parse ripgrep output: %w", err)
	}

	// Apply max results limit if needed
	if options.MaxResults > 0 && len(matches) > options.MaxResults {
		matches = matches[:options.MaxResults]
	}

	return SearchTextResult{
		AppName: appName,
		Pattern: pattern,
		Matches: matches,
		Total:   len(matches),
	}, nil
}

// parseRipgrepOutput parses the JSON output from ripgrep
func parseRipgrepOutput(output string, appDir string) ([]SearchMatch, error) {
	var matches []SearchMatch

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // Skip invalid JSON lines
		}

		// We're only interested in "match" type messages
		if msgType, ok := msg["type"].(string); ok && msgType == "match" {
			if data, ok := msg["data"].(map[string]interface{}); ok {
				match := SearchMatch{}

				// Extract file path
				if pathData, ok := data["path"].(map[string]interface{}); ok {
					if text, ok := pathData["text"].(string); ok {
						relPath, err := filepath.Rel(appDir, text)
						if err == nil {
							match.FilePath = relPath
						} else {
							match.FilePath = text
						}
					}
				}

				// Extract line number
				if lineNum, ok := data["line_number"].(float64); ok {
					match.LineNumber = int(lineNum)
				}

				// Extract line text and match
				if lines, ok := data["lines"].(map[string]interface{}); ok {
					if text, ok := lines["text"].(string); ok {
						match.LineText = strings.TrimRight(text, "\r\n")
					}
				}

				// Extract the actual match text
				if submatches, ok := data["submatches"].([]interface{}); ok && len(submatches) > 0 {
					if submatch, ok := submatches[0].(map[string]interface{}); ok {
						if matchData, ok := submatch["match"].(map[string]interface{}); ok {
							if text, ok := matchData["text"].(string); ok {
								match.Match = text
							}
						}
					}
				}

				matches = append(matches, match)
			}
		}
	}

	return matches, nil
}

// getRipgrepPath returns the path to the ripgrep binary
func getRipgrepPath() (string, error) {
	// Determine the platform-specific binary name
	var binaryName string
	switch runtime.GOOS {
	case "windows":
		binaryName = "rg.exe"
	default:
		binaryName = "rg"
	}

	// Get executable directory
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// For Windows ARM64, use the amd64 binary (emulation)
	arch := runtime.GOARCH
	if runtime.GOOS == "windows" && runtime.GOARCH == "arm64" {
		arch = "amd64"
	}

	// Look for ripgrep in order of preference:
	// 1. Bundled with executable (for release distributions)
	// 2. System PATH
	// 3. Development third-party directory
	searchPaths := []string{
		filepath.Join(execDir, binaryName), // Same directory as executable
		"",                                 // Empty string signals to check PATH
		filepath.Join(execDir, "third-party", "ripgrep", fmt.Sprintf("%s-%s", arch, runtime.GOOS), binaryName),
	}

	for _, path := range searchPaths {
		if path == "" {
			// Check system PATH
			if systemPath, err := exec.LookPath("rg"); err == nil {
				return systemPath, nil
			}
		} else {
			// Check specific path
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", errors.New("ripgrep is required but not found. Please install ripgrep: https://github.com/BurntSushi/ripgrep#installation")
}

// CLI
func SearchTextCli() error {
	args := os.Args[3:]

	// Check for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printSearchTextHelp()
			return nil
		}
	}

	var appName, pattern string
	var options SearchTextOptions
	options.MaxResults = 100 // Default limit

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
				pattern = args[i+1]
				i++
			} else {
				return errors.New("--pattern requires a value")
			}
		case "--case-sensitive":
			options.CaseSensitive = true
		case "--whole-word":
			options.WholeWord = true
		case "--file-pattern":
			if i+1 < len(args) {
				options.FilePattern = args[i+1]
				i++
			} else {
				return errors.New("--file-pattern requires a value")
			}
		case "--max-results":
			if i+1 < len(args) {
				var err error
				_, err = fmt.Sscanf(args[i+1], "%d", &options.MaxResults)
				if err != nil {
					return fmt.Errorf("--max-results must be a number: %w", err)
				}
				i++
			} else {
				return errors.New("--max-results requires a value")
			}
		case "--include-hidden":
			options.IncludeHidden = true
		default:
			if strings.HasPrefix(args[i], "--") {
				return fmt.Errorf("unknown option: %s\nRun 'layered-code tool search_text --help' for usage", args[i])
			}
		}
	}

	if appName == "" {
		return errors.New("--app-name is required")
	}
	if pattern == "" {
		return errors.New("--pattern is required")
	}

	result, err := SearchText(appName, pattern, options)
	if err != nil {
		return err
	}

	// Output results
	fmt.Printf("App: %s\nPattern: %s\nTotal matches: %d\n\n", result.AppName, result.Pattern, result.Total)

	for _, match := range result.Matches {
		fmt.Printf("%s:%d: %s\n", match.FilePath, match.LineNumber, match.LineText)
	}

	return nil
}

func printSearchTextHelp() {
	fmt.Println("Usage: layered-code tool search_text [options]")
	fmt.Println()
	fmt.Println("Search for text patterns in files within an application directory using ripgrep")
	fmt.Println()
	fmt.Println("Required options:")
	fmt.Println("  --app-name <name>        Name of the app directory")
	fmt.Println("  --pattern <pattern>      Search pattern (supports regular expressions)")
	fmt.Println()
	fmt.Println("Optional options:")
	fmt.Println("  --case-sensitive         Perform case-sensitive search (default: case-insensitive)")
	fmt.Println("  --whole-word             Match whole words only")
	fmt.Println("  --file-pattern <glob>    Only search files matching this glob pattern")
	fmt.Println("  --max-results <n>        Maximum number of results to return (default: 100)")
	fmt.Println("  --include-hidden         Include hidden files and directories in search")
	fmt.Println("  --help, -h               Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Search for a simple pattern")
	fmt.Println("  layered-code tool search_text --app-name myapp --pattern 'TODO'")
	fmt.Println()
	fmt.Println("  # Case-sensitive search in specific file types")
	fmt.Println("  layered-code tool search_text --app-name myapp --pattern 'Config' --case-sensitive --file-pattern '*.go'")
	fmt.Println()
	fmt.Println("  # Search for whole words with limited results")
	fmt.Println("  layered-code tool search_text --app-name myapp --pattern 'test' --whole-word --max-results 50")
}

// MCP
func SearchTextMcp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		AppName       string `json:"app_name"`
		Pattern       string `json:"pattern"`
		CaseSensitive bool   `json:"case_sensitive"`
		WholeWord     bool   `json:"whole_word"`
		FilePattern   string `json:"file_pattern"`
		MaxResults    int    `json:"max_results"`
		IncludeHidden bool   `json:"include_hidden"`
	}

	if err := request.BindArguments(&args); err != nil {
		return nil, err
	}

	// Set default max results if not specified
	if args.MaxResults == 0 {
		args.MaxResults = 100
	}

	options := SearchTextOptions{
		CaseSensitive: args.CaseSensitive,
		WholeWord:     args.WholeWord,
		FilePattern:   args.FilePattern,
		MaxResults:    args.MaxResults,
		IncludeHidden: args.IncludeHidden,
	}

	result, err := SearchText(args.AppName, args.Pattern, options)
	if err != nil {
		return nil, err
	}

	content, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(content)), nil
}
