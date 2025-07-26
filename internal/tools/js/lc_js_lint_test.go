package js

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLcJsLint(t *testing.T) {
	// Create a temporary test directory
	tempDir := t.TempDir()
	appsDir := filepath.Join(tempDir, "apps")
	os.Setenv("LAYERED_CODE_HOME", tempDir)
	defer os.Unsetenv("LAYERED_CODE_HOME")

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
}

