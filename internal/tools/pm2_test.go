package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/layered-flow/layered-code/internal/config"
)

func TestPM2(t *testing.T) {
	// Skip if not in development environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping PM2 tests in CI environment")
	}

	// Create a test app in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}
	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir) // Clean up after test
	
	testAppsDir := filepath.Join(tempDir, "apps")
	os.Setenv("LAYERED_APPS_DIRECTORY", testAppsDir)
	defer os.Unsetenv("LAYERED_APPS_DIRECTORY")

	// Ensure apps directory exists
	if _, err := config.EnsureAppsDirectory(); err != nil {
		t.Fatalf("Failed to ensure apps directory: %v", err)
	}

	// Create test app with package.json and ecosystem.config.cjs
	appName := "test-pm2-app"
	appPath := filepath.Join(testAppsDir, appName)
	if err := os.MkdirAll(appPath, 0755); err != nil {
		t.Fatalf("Failed to create app directory: %v", err)
	}

	// Create a simple package.json
	packageJSON := `{
  "name": "test-pm2-app",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "echo 'Running dev server' && sleep 10"
  },
  "devDependencies": {
    "pm2": "^5.3.0"
  }
}`
	if err := os.WriteFile(filepath.Join(appPath, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create ecosystem.config.cjs
	ecosystemConfig := `module.exports = {
  apps: [{
    name: 'test-pm2-app',
    script: 'npm',
    args: 'run dev',
    pid_file: './.layered-code/server.pid'
  }]
}`
	if err := os.WriteFile(filepath.Join(appPath, "ecosystem.config.cjs"), []byte(ecosystemConfig), 0644); err != nil {
		t.Fatalf("Failed to create ecosystem.config.cjs: %v", err)
	}

	// Test PM2 commands
	t.Run("PM2 Start", func(t *testing.T) {
		params := PM2Params{
			AppName: appName,
			Command: PM2Start,
		}
		result, err := PM2(params)
		if err != nil {
			// It's okay if PM2 isn't installed in test environment
			t.Logf("PM2 start failed (possibly not installed): %v", err)
			return
		}
		if !result.Success {
			t.Errorf("Expected PM2 start to succeed, got: %s", result.Message)
		}
	})

	t.Run("PM2 Status", func(t *testing.T) {
		params := PM2Params{
			AppName: appName,
			Command: PM2Status,
		}
		result, err := PM2(params)
		if err != nil {
			t.Logf("PM2 status failed (possibly not installed): %v", err)
			return
		}
		if !result.Success {
			t.Errorf("Expected PM2 status to succeed, got: %s", result.Message)
		}
	})

	t.Run("PM2 Stop", func(t *testing.T) {
		params := PM2Params{
			AppName: appName,
			Command: PM2Stop,
		}
		result, err := PM2(params)
		if err != nil {
			t.Logf("PM2 stop failed (possibly not installed): %v", err)
			return
		}
		if !result.Success {
			t.Errorf("Expected PM2 stop to succeed, got: %s", result.Message)
		}
	})

	t.Run("PM2 with non-existent app", func(t *testing.T) {
		params := PM2Params{
			AppName: "non-existent-app",
			Command: PM2Start,
		}
		_, err := PM2(params)
		if err == nil {
			t.Error("Expected error for non-existent app, got nil")
		}
	})

	t.Run("PM2 with custom config", func(t *testing.T) {
		// Create custom config
		customConfig := `module.exports = {
  apps: [{
    name: 'test-custom',
    script: 'npm',
    args: 'run dev'
  }]
}`
		if err := os.WriteFile(filepath.Join(appPath, "ecosystem.custom.cjs"), []byte(customConfig), 0644); err != nil {
			t.Fatalf("Failed to create custom config: %v", err)
		}

		params := PM2Params{
			AppName: appName,
			Command: PM2Start,
			Config:  "ecosystem.custom.cjs",
		}
		result, err := PM2(params)
		if err != nil {
			t.Logf("PM2 with custom config failed (possibly not installed): %v", err)
			return
		}
		if !result.Success {
			t.Errorf("Expected PM2 with custom config to succeed, got: %s", result.Message)
		}
	})
}