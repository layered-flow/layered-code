package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildApp(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	
	t.Run("successful build", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		// Create a test app with package.json
		appName := "buildapp"
		appPath := filepath.Join(appsDir, appName)
		os.MkdirAll(appPath, 0755)

		// Create a minimal package.json with build script
		packageJSON := map[string]interface{}{
			"name": appName,
			"scripts": map[string]string{
				"build": "echo 'Building project...'",
			},
		}
		packageData, _ := json.MarshalIndent(packageJSON, "", "  ")
		os.WriteFile(filepath.Join(appPath, "package.json"), packageData, 0644)

		params := BuildAppParams{AppName: appName}
		result, err := BuildApp(params)
		if err != nil {
			t.Fatalf("BuildApp() failed: %v", err)
		}

		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "Successfully built app")
		assert.Contains(t, result.Output, "Building project...")
	})

	t.Run("app not found", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		params := BuildAppParams{AppName: "nonexistentapp"}
		result, err := BuildApp(params)
		
		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "does not exist")
	})

	t.Run("missing package.json", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		// Create app without package.json
		appName := "nopkgjson"
		appPath := filepath.Join(appsDir, appName)
		os.MkdirAll(appPath, 0755)

		params := BuildAppParams{AppName: appName}
		result, err := BuildApp(params)
		
		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "package.json not found")
	})

	t.Run("missing build script", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		// Create app with package.json but no build script
		appName := "nobuildscript"
		appPath := filepath.Join(appsDir, appName)
		os.MkdirAll(appPath, 0755)

		packageJSON := map[string]interface{}{
			"name": appName,
			"scripts": map[string]string{
				"test": "echo 'Testing...'",
			},
		}
		packageData, _ := json.MarshalIndent(packageJSON, "", "  ")
		os.WriteFile(filepath.Join(appPath, "package.json"), packageData, 0644)

		params := BuildAppParams{AppName: appName}
		result, err := BuildApp(params)
		
		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "no 'build' script found")
	})

	t.Run("invalid app name", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		params := BuildAppParams{AppName: "invalid/name"}
		result, err := BuildApp(params)
		
		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "invalid characters")
	})
}

// TestBuildAppCli tests the CLI interface
func TestBuildAppCli(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	t.Run("missing arguments", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "build_app"}
		err := BuildAppCli()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires exactly 1 argument")
	})

	t.Run("too many arguments", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "build_app", "app1", "extra"}
		err := BuildAppCli()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires exactly 1 argument")
	})
}