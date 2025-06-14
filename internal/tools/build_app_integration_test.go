// +build integration

package tools

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAppIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	t.Run("vite app creates dist directory", func(t *testing.T) {
		// Create a Vite app
		appName := "vite-build-test"
		createParams := CreateAppParams{
			AppName: appName,
		}
		
		createResult, err := CreateApp(createParams)
		require.NoError(t, err)
		assert.True(t, createResult.Success)

		appPath := filepath.Join(appsDir, appName)

		// Install dependencies
		installParams := NpmInstallParams{
			AppName: appName,
		}
		
		t.Log("Installing dependencies...")
		installResult, err := NpmInstall(installParams)
		require.NoError(t, err)
		assert.True(t, installResult.Success)

		// Build the app
		buildParams := BuildAppParams{
			AppName: appName,
		}
		
		t.Log("Building app...")
		startTime := time.Now()
		buildResult, err := BuildApp(buildParams)
		require.NoError(t, err)
		assert.True(t, buildResult.Success)
		t.Logf("Build completed in %s", time.Since(startTime))
		t.Logf("Build output:\n%s", buildResult.Output)

		// Verify dist directory was created
		distPath := filepath.Join(appPath, "dist")
		distInfo, err := os.Stat(distPath)
		require.NoError(t, err, "dist directory should exist")
		assert.True(t, distInfo.IsDir(), "dist should be a directory")

		// Verify dist contains built files
		entries, err := os.ReadDir(distPath)
		require.NoError(t, err)
		assert.Greater(t, len(entries), 0, "dist directory should contain files")

		// Check for index.html in dist
		indexPath := filepath.Join(distPath, "index.html")
		_, err = os.Stat(indexPath)
		assert.NoError(t, err, "dist/index.html should exist")

		// Verify build directory does NOT exist
		buildPath := filepath.Join(appPath, "build")
		_, err = os.Stat(buildPath)
		assert.True(t, os.IsNotExist(err), "build directory should not exist")
	})
}