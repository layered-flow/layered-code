package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestSearchText tests the core SearchText functionality
func TestSearchText(t *testing.T) {
	// Skip if ripgrep is not available
	if _, err := getRipgrepPath(); err != nil {
		t.Skip("Skipping test: ripgrep not available")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	srcDir := filepath.Join(appDir, "src")
	os.MkdirAll(srcDir, 0755)

	// Create test files with various content
	testFiles := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	// TODO: Add more features
}`,
		"utils.go": `package main

func HelperFunction() {
	// Helper function implementation
	// TODO: Implement this
}`,
		"src/config.go": `package src

type Config struct {
	Name string
	Port int
}

func LoadConfig() *Config {
	// TODO: Load from file
	return &Config{Name: "test", Port: 8080}
}`,
		"README.md": `# Test Application

This is a test application for searching.

## TODO List
- Add documentation
- Write tests
- Deploy to production`,
		".hidden": `This is a hidden file with TODO items`,
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(appDir, path)
		dir := filepath.Dir(fullPath)
		if dir != appDir {
			os.MkdirAll(dir, 0755)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	tests := []struct {
		name         string
		pattern      string
		options      SearchTextOptions
		wantMinCount int
		wantMaxCount int
		checkMatches func(t *testing.T, matches []SearchMatch)
	}{
		{
			name:         "simple pattern search",
			pattern:      "TODO",
			options:      SearchTextOptions{},
			wantMinCount: 4, // Should find TODO in multiple files
		},
		{
			name:    "case sensitive search",
			pattern: "todo",
			options: SearchTextOptions{
				CaseSensitive: true,
			},
			wantMinCount: 0,
			wantMaxCount: 0, // Should find nothing (all TODOs are uppercase)
		},
		{
			name:    "case insensitive search",
			pattern: "todo",
			options: SearchTextOptions{
				CaseSensitive: false,
			},
			wantMinCount: 4,
		},
		{
			name:    "whole word search",
			pattern: "Config",
			options: SearchTextOptions{
				WholeWord: true,
			},
			wantMinCount: 2, // Should match Config but not LoadConfig
		},
		{
			name:    "file pattern filter",
			pattern: "TODO",
			options: SearchTextOptions{
				FilePattern: "*.go",
			},
			wantMinCount: 3,
			checkMatches: func(t *testing.T, matches []SearchMatch) {
				for _, match := range matches {
					if !strings.HasSuffix(match.FilePath, ".go") {
						t.Errorf("Expected only .go files, got: %s", match.FilePath)
					}
				}
			},
		},
		{
			name:    "max results limit",
			pattern: "TODO",
			options: SearchTextOptions{
				MaxResults: 2,
			},
			wantMinCount: 2,
			wantMaxCount: 2,
		},
		{
			name:    "include hidden files",
			pattern: "TODO",
			options: SearchTextOptions{
				IncludeHidden: true,
			},
			wantMinCount: 5, // Should include .hidden file
		},
		{
			name:         "no matches",
			pattern:      "NONEXISTENT",
			options:      SearchTextOptions{},
			wantMinCount: 0,
			wantMaxCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SearchText("testapp", tt.pattern, tt.options)
			if err != nil {
				t.Fatalf("SearchText() failed: %v", err)
			}

			if result.AppName != "testapp" {
				t.Errorf("AppName = %s; want testapp", result.AppName)
			}
			if result.Pattern != tt.pattern {
				t.Errorf("Pattern = %s; want %s", result.Pattern, tt.pattern)
			}

			if tt.wantMinCount > 0 && len(result.Matches) < tt.wantMinCount {
				t.Errorf("Got %d matches; want at least %d", len(result.Matches), tt.wantMinCount)
			}
			if tt.wantMaxCount > 0 && len(result.Matches) > tt.wantMaxCount {
				t.Errorf("Got %d matches; want at most %d", len(result.Matches), tt.wantMaxCount)
			}

			// Verify each match has required fields
			for _, match := range result.Matches {
				if match.FilePath == "" {
					t.Error("Match has empty FilePath")
				}
				if match.LineNumber <= 0 {
					t.Errorf("Match has invalid LineNumber: %d", match.LineNumber)
				}
				if match.LineText == "" {
					t.Error("Match has empty LineText")
				}
			}

			// Run custom match checks if provided
			if tt.checkMatches != nil {
				tt.checkMatches(t, result.Matches)
			}
		})
	}
}

// TestSearchTextErrors tests error conditions
func TestSearchTextErrors(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	os.MkdirAll(appsDir, 0755)
	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	tests := []struct {
		name    string
		appName string
		pattern string
		wantErr string
	}{
		{
			name:    "empty app name",
			appName: "",
			pattern: "test",
			wantErr: "app_name is required",
		},
		{
			name:    "empty pattern",
			appName: "testapp",
			pattern: "",
			wantErr: "pattern is required",
		},
		{
			name:    "non-existent app",
			appName: "nonexistent",
			pattern: "test",
			wantErr: "app 'nonexistent' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SearchText(tt.appName, tt.pattern, SearchTextOptions{})
			if err == nil {
				t.Fatal("Expected error but got none")
			}
			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("Error = %v; want error containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestSearchTextCli tests the CLI interface
func TestSearchTextCli(t *testing.T) {
	// Skip if ripgrep is not available
	if _, err := getRipgrepPath(); err != nil {
		t.Skip("Skipping test: ripgrep not available")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)

	testContent := "// TODO: Test content"
	os.WriteFile(filepath.Join(appDir, "test.go"), []byte(testContent), 0644)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help flag",
			args:    []string{"", "", "", "--help"},
			wantErr: false,
		},
		{
			name:    "missing app name",
			args:    []string{"", "", "", "--pattern", "TODO"},
			wantErr: true,
		},
		{
			name:    "missing pattern",
			args:    []string{"", "", "", "--app-name", "testapp"},
			wantErr: true,
		},
		{
			name:    "successful search",
			args:    []string{"", "", "", "--app-name", "testapp", "--pattern", "TODO"},
			wantErr: false,
		},
		{
			name:    "with options",
			args:    []string{"", "", "", "--app-name", "testapp", "--pattern", "TODO", "--case-sensitive", "--max-results", "5"},
			wantErr: false,
		},
		{
			name:    "invalid max results",
			args:    []string{"", "", "", "--app-name", "testapp", "--pattern", "TODO", "--max-results", "invalid"},
			wantErr: true,
		},
		{
			name:    "unknown option",
			args:    []string{"", "", "", "--unknown-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = tt.args
			err := SearchTextCli()

			if (err != nil) != tt.wantErr {
				t.Errorf("SearchTextCli() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSearchTextMcp tests the MCP interface
func TestSearchTextMcp(t *testing.T) {
	// Skip if ripgrep is not available
	if _, err := getRipgrepPath(); err != nil {
		t.Skip("Skipping test: ripgrep not available")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	appDir := filepath.Join(appsDir, "testapp")
	os.MkdirAll(appDir, 0755)

	// Create a test file
	testContent := `package main
// TODO: Implement feature`
	os.WriteFile(filepath.Join(appDir, "main.go"), []byte(testContent), 0644)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	t.Run("successful search via MCP", func(t *testing.T) {
		args := map[string]interface{}{
			"app_name": "testapp",
			"pattern":  "TODO",
		}

		request := mcp.CallToolRequest{}
		data, _ := json.Marshal(args)
		json.Unmarshal(data, &request.Params.Arguments)
		result, err := SearchTextMcp(context.Background(), request)
		if err != nil {
			t.Fatalf("SearchTextMcp() failed: %v", err)
		}

		// The result should contain JSON with matches
		if result == nil || len(result.Content) == 0 {
			t.Error("Expected non-empty result")
			return
		}

		var textContent []mcp.TextContent
		for _, c := range result.Content {
			if tc, ok := c.(mcp.TextContent); ok {
				textContent = append(textContent, tc)
			}
		}

		if len(textContent) == 0 {
			t.Error("Expected text content in result")
			return
		}

		if !containsString(textContent[0].Text, "TODO") {
			t.Errorf("Result does not contain expected pattern")
		}
	})

	t.Run("MCP with options", func(t *testing.T) {
		args := map[string]interface{}{
			"app_name":       "testapp",
			"pattern":        "TODO",
			"case_sensitive": true,
			"file_pattern":   "*.go",
			"max_results":    10,
		}

		request := mcp.CallToolRequest{}
		data, _ := json.Marshal(args)
		json.Unmarshal(data, &request.Params.Arguments)
		result, err := SearchTextMcp(context.Background(), request)
		if err != nil {
			t.Fatalf("SearchTextMcp() failed: %v", err)
		}

		if result == nil || len(result.Content) == 0 {
			t.Error("Expected non-empty result")
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(s) > len(substr) && containsString(s[1:len(s)-1], substr)))
}
