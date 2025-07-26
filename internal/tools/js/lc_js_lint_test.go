package js

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestLcJsLint(t *testing.T) {
	// Create a temporary test directory within home directory for security validation
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	appsDir := filepath.Join(tempDir, "apps")
	os.Setenv("LAYERED_APPS_DIRECTORY", appsDir)
	defer os.Unsetenv("LAYERED_APPS_DIRECTORY")

	// Create apps directory
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test app
	testApp := filepath.Join(appsDir, "testapp")
	if err := os.MkdirAll(testApp, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a simple JS file
	jsFile := filepath.Join(testApp, "test.js")
	jsContent := `function hello() {
    console.log("hello");
}`
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create package.json for ESLint
	packageJSON := filepath.Join(testApp, "package.json")
	packageContent := `{
  "name": "testapp",
  "version": "1.0.0"
}`
	if err := os.WriteFile(packageJSON, []byte(packageContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("single file", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"test.js"},
		}

		result, err := LcJsLint(params)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should succeed even if ESLint is not installed - graceful handling
		if !result.Success && result.Message != "" {
			// Check for expected messages
			if !strings.Contains(result.Message, "ESLint") && !strings.Contains(result.Message, "App directory does not exist") {
				t.Errorf("Unexpected message: %s", result.Message)
			}
		}
	})

	t.Run("multiple files as array", func(t *testing.T) {
		// Create another JS file
		jsFile2 := filepath.Join(testApp, "test2.js")
		if err := os.WriteFile(jsFile2, []byte(jsContent), 0644); err != nil {
			t.Fatal(err)
		}

		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"test.js", "test2.js"},
		}

		result, err := LcJsLint(params)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should succeed even if ESLint is not installed - graceful handling
		if !result.Success && result.Message != "" {
			// Check for expected messages
			if !strings.Contains(result.Message, "ESLint") && !strings.Contains(result.Message, "App directory does not exist") {
				t.Errorf("Unexpected message: %s", result.Message)
			}
		}
	})

	t.Run("glob pattern", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"*.js"},
		}

		result, err := LcJsLint(params)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Should succeed even if ESLint is not installed - graceful handling
		if !result.Success && result.Message != "" {
			// Check for expected messages
			if !strings.Contains(result.Message, "ESLint") && !strings.Contains(result.Message, "App directory does not exist") {
				t.Errorf("Unexpected message: %s", result.Message)
			}
		}
	})

	t.Run("empty files array", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{},
		}

		_, err := LcJsLint(params)
		if err == nil {
			t.Error("Expected error for empty files array")
		}
		if !strings.Contains(err.Error(), "files parameter cannot be empty") {
			t.Errorf("Expected 'files parameter cannot be empty' error, got: %v", err)
		}
	})

	t.Run("empty string in files array", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"test.js", ""},
		}

		_, err := LcJsLint(params)
		if err == nil {
			t.Error("Expected error for empty string in files array")
		}
		if !strings.Contains(err.Error(), "file patterns cannot be empty") {
			t.Errorf("Expected 'file patterns cannot be empty' error, got: %v", err)
		}
	})

	t.Run("path traversal in files", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"../../../etc/passwd"},
		}

		_, err := LcJsLint(params)
		if err == nil {
			t.Error("Expected error for path traversal")
		}
		if !strings.Contains(err.Error(), "file pattern cannot contain '..':") {
			t.Errorf("Expected path traversal error, got: %v", err)
		}
	})

	t.Run("absolute path in files", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"/etc/passwd"},
		}

		_, err := LcJsLint(params)
		if err == nil {
			t.Error("Expected error for absolute path")
		}
		if !strings.Contains(err.Error(), "absolute paths are not allowed:") {
			t.Errorf("Expected absolute path error, got: %v", err)
		}
	})

	t.Run("path traversal in config", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"test.js"},
			Config:  "../../../etc/passwd",
		}

		_, err := LcJsLint(params)
		if err == nil {
			t.Error("Expected error for config path traversal")
		} else if !strings.Contains(err.Error(), "config path cannot contain '..'") {
			t.Errorf("Expected config path traversal error, got: %v", err)
		}
	})

	t.Run("absolute path in config", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"test.js"},
			Config:  "/etc/passwd",
		}

		_, err := LcJsLint(params)
		if err == nil {
			t.Error("Expected error for absolute config path")
		} else if !strings.Contains(err.Error(), "config path cannot be absolute:") {
			t.Errorf("Expected absolute config path error, got: %v", err)
		}
	})

	t.Run("non-existent app directory", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "nonexistent",
			Files:   []string{"test.js"},
		}

		result, err := LcJsLint(params)
		if err != nil {
			t.Errorf("Expected graceful error handling, got error: %v", err)
		}

		if !result.Success {
			if !strings.Contains(result.Message, "App directory does not exist") {
				t.Errorf("Expected 'App directory does not exist' message, got: %s", result.Message)
			}
		}
	})

	t.Run("files with spaces in array", func(t *testing.T) {
		// Create files with spaces
		fileWithSpace := filepath.Join(testApp, "file with space.js")
		if err := os.WriteFile(fileWithSpace, []byte(jsContent), 0644); err != nil {
			t.Fatal(err)
		}

		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"test.js", "file with space.js"},
		}

		result, err := LcJsLint(params)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check that the command includes proper handling of spaces
		if result.Command != "" && strings.Contains(result.Command, "file with space.js") {
			// Command should handle spaces properly
			t.Logf("Command: %s", result.Command)
		}
	})

	t.Run("with fix flag", func(t *testing.T) {
		params := LcJsLintParams{
			AppName: "testapp",
			Files:   []string{"test.js"},
			Fix:     true,
		}

		result, err := LcJsLint(params)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check that the command includes --fix
		if result.Command != "" && !strings.Contains(result.Command, "--fix") {
			t.Errorf("Expected command to contain --fix flag, got: %s", result.Command)
		}
	})
}

func TestLcJsLintCli(t *testing.T) {
	// Create a temporary test directory within home directory for security validation
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := filepath.Join(homeDir, ".layered-test-cli-"+t.Name())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	appsDir := filepath.Join(tempDir, "apps")
	os.Setenv("LAYERED_APPS_DIRECTORY", appsDir)
	defer os.Unsetenv("LAYERED_APPS_DIRECTORY")

	// Create apps directory
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test app
	testApp := filepath.Join(appsDir, "testapp")
	if err := os.MkdirAll(testApp, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a simple JS file
	jsFile := filepath.Join(testApp, "test.js")
	jsContent := `function hello() {
    console.log("hello");
}`
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original os.Args
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	t.Run("CLI with --fix flag should be allowed", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "--fix", "test.js"}
		
		// This should not error at CLI parsing level
		err := LcJsLintCli()
		// The actual ESLint execution might fail if ESLint is not installed, but --fix should be accepted
		if err != nil && strings.Contains(err.Error(), "invalid option: --fix") {
			t.Errorf("Should not get 'invalid option: --fix' error, got: %v", err)
		}
	})

	t.Run("CLI with --no-eslintrc flag should be rejected", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "test.js", "--no-eslintrc"}
		
		err := LcJsLintCli()
		if err == nil {
			t.Error("Expected error for --no-eslintrc flag")
		}
		if !strings.Contains(err.Error(), "invalid option: --no-eslintrc") {
			t.Errorf("Expected 'invalid option: --no-eslintrc' error, got: %v", err)
		}
	})

	t.Run("CLI with invalid flag after valid ones", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "--fix", "test.js", "--quiet"}
		
		err := LcJsLintCli()
		if err == nil {
			t.Error("Expected error for --quiet flag")
		}
		if !strings.Contains(err.Error(), "invalid option: --quiet") {
			t.Errorf("Expected 'invalid option: --quiet' error, got: %v", err)
		}
	})

	t.Run("CLI with valid --config flag", func(t *testing.T) {
		// Create a config file
		configFile := filepath.Join(testApp, ".eslintrc.json")
		configContent := `{
  "rules": {
    "semi": ["error", "always"]
  }
}`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "--config=.eslintrc.json", "test.js"}
		
		// This should not error at CLI parsing level
		err := LcJsLintCli()
		// The actual ESLint execution might fail if ESLint is not installed, but CLI parsing should succeed
		if err != nil && strings.Contains(err.Error(), "invalid option") {
			t.Errorf("Should not get 'invalid option' error for valid --config flag, got: %v", err)
		}
	})

	t.Run("CLI with file starting with --", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "--test.js"}
		
		err := LcJsLintCli()
		if err == nil {
			t.Error("Expected error for file starting with --")
		}
		if !strings.Contains(err.Error(), "invalid option: --test.js") {
			t.Errorf("Expected 'invalid option: --test.js' error, got: %v", err)
		}
	})

	t.Run("CLI with absolute path file", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "/etc/passwd"}
		
		err := LcJsLintCli()
		if err == nil {
			t.Error("Expected error for absolute path")
		}
		if !strings.Contains(err.Error(), "absolute paths are not allowed:") {
			t.Errorf("Expected 'absolute paths are not allowed' error, got: %v", err)
		}
	})

	t.Run("CLI with absolute config path", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "--config=/etc/eslintrc", "test.js"}
		
		err := LcJsLintCli()
		if err == nil {
			t.Error("Expected error for absolute config path")
		}
		if !strings.Contains(err.Error(), "config path cannot be absolute:") {
			t.Errorf("Expected 'config path cannot be absolute' error, got: %v", err)
		}
	})

	t.Run("CLI with glob patterns", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "src/**/*.js", "tests/**/*.js"}
		
		// This should not error at CLI parsing level
		err := LcJsLintCli()
		// The actual ESLint execution might fail if ESLint is not installed, but CLI parsing should succeed
		if err != nil && strings.Contains(err.Error(), "invalid option") {
			t.Errorf("Should not get 'invalid option' error for glob patterns, got: %v", err)
		}
	})

	t.Run("CLI with config containing path traversal", func(t *testing.T) {
		// Debug: log the exact arguments
		testArgs := []string{"layered-code", "tool", "lc_js_lint", "testapp", "--config=../../etc/passwd", "test.js"}
		t.Logf("Setting os.Args to: %v", testArgs)
		os.Args = testArgs
		
		err := LcJsLintCli()
		if err == nil {
			t.Fatal("Expected error for config with path traversal, but got nil")
		}
		if !strings.Contains(err.Error(), "config path cannot contain '..'") {
			t.Errorf("Expected 'config path cannot contain ..' error, but got: %v", err)
		}
	})

	t.Run("CLI with both --fix and --config flags", func(t *testing.T) {
		// Create a config file
		configFile := filepath.Join(testApp, ".eslintrc.json")
		configContent := `{
  "rules": {
    "semi": ["error", "always"]
  }
}`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		os.Args = []string{"layered-code", "tool", "lc_js_lint", "testapp", "--fix", "--config=.eslintrc.json", "test.js"}
		
		// This should not error at CLI parsing level
		err := LcJsLintCli()
		// The actual ESLint execution might fail if ESLint is not installed, but CLI parsing should succeed
		if err != nil && strings.Contains(err.Error(), "invalid option") {
			t.Errorf("Should not get 'invalid option' error for valid flags, got: %v", err)
		}
	})
}

func TestLcJsLintMcp(t *testing.T) {
	// Import required packages for MCP testing
	ctx := context.Background()
	
	// Create a temporary test directory within home directory for security validation
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := filepath.Join(homeDir, ".layered-test-mcp-"+t.Name())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	appsDir := filepath.Join(tempDir, "apps")
	os.Setenv("LAYERED_APPS_DIRECTORY", appsDir)
	defer os.Unsetenv("LAYERED_APPS_DIRECTORY")

	// Create apps directory
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test app
	testApp := filepath.Join(appsDir, "testapp")
	if err := os.MkdirAll(testApp, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a simple JS file
	jsFile := filepath.Join(testApp, "test.js")
	jsContent := `function hello() {
    console.log("hello");
}`
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("MCP with fix flag", func(t *testing.T) {
		// Create a mock MCP request
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"app_name": "testapp",
			"files":    []string{"test.js"},
			"fix":      true,
		}

		result, err := LcJsLintMcp(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Check that result contains JSON data
		if result == nil || len(result.Content) == 0 {
			t.Error("Expected result with content")
		} else {
			// Verify it's valid JSON
			var jsonResult LcJsLintResult
			textContent := result.Content[0].(mcp.TextContent).Text
			if err := json.Unmarshal([]byte(textContent), &jsonResult); err != nil {
				t.Errorf("Expected valid JSON result, got error: %v", err)
			}
			// Check that command includes --fix
			if !strings.Contains(jsonResult.Command, "--fix") {
				t.Errorf("Expected command to contain --fix flag, got: %s", jsonResult.Command)
			}
		}
	})

	t.Run("MCP with config", func(t *testing.T) {
		// Create a config file
		configFile := filepath.Join(testApp, ".eslintrc.json")
		configContent := `{
  "rules": {
    "semi": ["error", "always"]
  }
}`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatal(err)
		}

		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]interface{}{
			"app_name": "testapp",
			"files":    []string{"test.js"},
			"config":   ".eslintrc.json",
		}

		result, err := LcJsLintMcp(ctx, request)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != nil && len(result.Content) > 0 {
			var jsonResult LcJsLintResult
			textContent := result.Content[0].(mcp.TextContent).Text
			if err := json.Unmarshal([]byte(textContent), &jsonResult); err == nil {
				// Check that command includes config
				if !strings.Contains(jsonResult.Command, "--config") {
					t.Errorf("Expected command to contain --config flag, got: %s", jsonResult.Command)
				}
			}
		}
	})
}

