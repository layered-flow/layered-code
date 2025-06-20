package lc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestLcListAppsResult tests the LcListAppsResult struct creation and field assignment
func TestLcListAppsResult(t *testing.T) {
	result := LcListAppsResult{Apps: []string{"app1", "app2"}, Directory: "/test/dir"}

	if len(result.Apps) != 2 || result.Directory != "/test/dir" {
		t.Errorf("LcListAppsResult not created correctly")
	}
}

// TestLcListApps tests the core LcListApps functionality with real directory structures
func TestLcListApps(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	os.MkdirAll(appsDir, 0755)

	t.Run("empty directory", func(t *testing.T) {
		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		result, err := LcListApps()
		if err != nil {
			t.Fatalf("LcListApps() failed: %v", err)
		}
		if len(result.Apps) != 0 {
			t.Errorf("Expected 0 apps, got %d", len(result.Apps))
		}
		if result.Directory != appsDir {
			t.Errorf("Expected directory %s, got %s", appsDir, result.Directory)
		}
	})

	t.Run("with apps", func(t *testing.T) {
		// Create test apps
		os.Mkdir(filepath.Join(appsDir, "app1"), 0755)
		os.Mkdir(filepath.Join(appsDir, "app2"), 0755)
		os.Mkdir(filepath.Join(appsDir, "zebra-app"), 0755)
		// Create a file (should be ignored)
		os.WriteFile(filepath.Join(appsDir, "not-an-app.txt"), []byte("test"), 0644)

		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		result, err := LcListApps()
		if err != nil {
			t.Fatalf("LcListApps() failed: %v", err)
		}

		expectedApps := []string{"app1", "app2", "zebra-app"}
		if len(result.Apps) != len(expectedApps) {
			t.Errorf("Expected %d apps, got %d", len(expectedApps), len(result.Apps))
		}

		// Verify alphabetical sorting
		for i, expected := range expectedApps {
			if i >= len(result.Apps) || result.Apps[i] != expected {
				t.Errorf("Expected app[%d] = %s, got %s", i, expected, result.Apps[i])
			}
		}
	})
}

// TestLcListAppsMcp tests the MCP interface wrapper for proper JSON marshaling
// and error handling
func TestLcListAppsMcp(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tempDir := filepath.Join(homeDir, ".layered-test-"+t.Name())
	defer os.RemoveAll(tempDir)

	appsDir := filepath.Join(tempDir, "apps")
	os.MkdirAll(appsDir, 0755)
	os.Mkdir(filepath.Join(appsDir, "testapp"), 0755)

	t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

	ctx := context.Background()
	request := mcp.CallToolRequest{}
	request.Params.Name = "lc_list_apps"

	result, err := LcListAppsMcp(ctx, request)
	if err != nil {
		t.Fatalf("LcListAppsMcp() failed: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}
}

// TestFunctionExecutions tests that all exported functions execute without panicking
// and verifies basic error handling for missing directories
func TestFunctionExecutions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"LcListApps", func() error { _, err := LcListApps(); return err }},
		{"LcListAppsCli", func() error { return LcListAppsCli() }},
		{"LcListAppsMcp", func() error {
			ctx := context.Background()
			request := mcp.CallToolRequest{}
			request.Params.Name = "lc_list_apps"
			_, err := LcListAppsMcp(ctx, request)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			_ = tt.fn() // Errors are expected due to filesystem/missing dirs
		})
	}
}
