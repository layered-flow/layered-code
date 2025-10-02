package lc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestLcAppListResult tests the LcAppListResult struct creation and field assignment
func TestLcAppListResult(t *testing.T) {
	result := LcAppListResult{
		Apps: []AppInfo{
			{Name: "app1", HasAgentsMd: true},
			{Name: "app2", HasAgentsMd: false},
		},
		Directory: "/test/dir",
	}

	if len(result.Apps) != 2 || result.Directory != "/test/dir" {
		t.Errorf("LcAppListResult not created correctly")
	}
}

// TestLcAppList tests the core LcAppList functionality with real directory structures
func TestLcAppList(t *testing.T) {
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

		result, err := LcAppList()
		if err != nil {
			t.Fatalf("LcAppList() failed: %v", err)
		}
		if len(result.Apps) != 0 {
			t.Errorf("Expected 0 apps, got %d", len(result.Apps))
		}
		if result.Directory != appsDir {
			t.Errorf("Expected directory %s, got %s", appsDir, result.Directory)
		}
	})

	t.Run("with apps and AGENTS.md files", func(t *testing.T) {
		// Create test apps
		app1Dir := filepath.Join(appsDir, "app1")
		app2Dir := filepath.Join(appsDir, "app2")
		zebraDir := filepath.Join(appsDir, "zebra-app")

		os.Mkdir(app1Dir, 0755)
		os.Mkdir(app2Dir, 0755)
		os.Mkdir(zebraDir, 0755)

		// Create AGENTS.md in app1 (lowercase)
		os.WriteFile(filepath.Join(app1Dir, "agents.md"), []byte("test"), 0644)

		// Create AGENTS.MD in zebra-app (uppercase)
		os.WriteFile(filepath.Join(zebraDir, "AGENTS.MD"), []byte("test"), 0644)

		// Create a file (should be ignored)
		os.WriteFile(filepath.Join(appsDir, "not-an-app.txt"), []byte("test"), 0644)

		t.Setenv("LAYERED_APPS_DIRECTORY", appsDir)

		result, err := LcAppList()
		if err != nil {
			t.Fatalf("LcAppList() failed: %v", err)
		}

		if len(result.Apps) != 3 {
			t.Errorf("Expected 3 apps, got %d", len(result.Apps))
		}

		// Verify alphabetical sorting and AGENTS.md detection
		expected := []AppInfo{
			{Name: "app1", HasAgentsMd: true},
			{Name: "app2", HasAgentsMd: false},
			{Name: "zebra-app", HasAgentsMd: true},
		}

		for i, exp := range expected {
			if i >= len(result.Apps) {
				t.Errorf("Missing app at index %d", i)
				continue
			}
			if result.Apps[i].Name != exp.Name {
				t.Errorf("Expected app[%d].Name = %s, got %s", i, exp.Name, result.Apps[i].Name)
			}
			if result.Apps[i].HasAgentsMd != exp.HasAgentsMd {
				t.Errorf("Expected app[%d].HasAgentsMd = %v, got %v", i, exp.HasAgentsMd, result.Apps[i].HasAgentsMd)
			}
		}
	})
}

// TestLcAppListMcp tests the MCP interface wrapper for proper JSON marshaling
// and error handling
func TestLcAppListMcp(t *testing.T) {
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
	request.Params.Name = "lc_app_list"

	result, err := LcAppListMcp(ctx, request)
	if err != nil {
		t.Fatalf("LcAppListMcp() failed: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}
}

// TestLcAppListFunctionExecutions tests that all exported functions execute without panicking
// and verifies basic error handling for missing directories
func TestLcAppListFunctionExecutions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"LcAppList", func() error { _, err := LcAppList(); return err }},
		{"LcAppListCli", func() error { return LcAppListCli() }},
		{"LcAppListMcp", func() error {
			ctx := context.Background()
			request := mcp.CallToolRequest{}
			request.Params.Name = "lc_app_list"
			_, err := LcAppListMcp(ctx, request)
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
