package cli

import (
	"os"
	"strings"
	"testing"
)

func TestRunTool(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "insufficient arguments",
			args:    []string{"layered-code", "tool"},
			wantErr: true,
		},
		{
			name:    "unknown subcommand",
			args:    []string{"layered-code", "tool", "unknown"},
			wantErr: true,
		},
		{
			name:    "lc_list_apps subcommand",
			args:    []string{"layered-code", "tool", "lc_list_apps"},
			wantErr: false, // Don't assert on error since it depends on external state
		},
		{
			name:    "vite_create_app subcommand",
			args:    []string{"layered-code", "tool", "vite_create_app", "test-app"},
			wantErr: false, // Don't assert on error since it depends on external state
		},
		{
			name:    "pnpm_install subcommand",
			args:    []string{"layered-code", "tool", "pnpm_install", "test-app"},
			wantErr: false, // Don't assert on error since it depends on external state
		},
		{
			name:    "pnpm_add subcommand",
			args:    []string{"layered-code", "tool", "pnpm_add", "test-app", "express"},
			wantErr: false, // Don't assert on error since it depends on external state
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			os.Args = tt.args
			err := RunTool()

			if tt.name == "lc_list_apps subcommand" || tt.name == "vite_create_app subcommand" || tt.name == "pnpm_install subcommand" || tt.name == "pnpm_add subcommand" {
				// For lc_list_apps, vite_create_app, pnpm_install, and pnpm_add, we just verify they don't panic and run the code path
				// Don't assert on error since they depend on external dependencies
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("RunTool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunToolErrorMessages(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("insufficient args error message", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool"}
		err := RunTool()
		if err == nil || !strings.Contains(err.Error(), "tool subcommand is required") {
			t.Errorf("Expected 'tool subcommand is required' error, got %v", err)
		}
	})

	t.Run("unknown tool error message", func(t *testing.T) {
		os.Args = []string{"layered-code", "tool", "invalid"}
		err := RunTool()
		if err == nil || !strings.Contains(err.Error(), "unknown tool: invalid") {
			t.Errorf("Expected 'unknown tool: invalid' error, got %v", err)
		}
		// Verify the error suggests running help
		if err == nil || !strings.Contains(err.Error(), "Run 'layered-code help' to see all available tools") {
			t.Errorf("Expected error to suggest running help, got %v", err)
		}
	})
}

func TestPrintUsage(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintUsage() panicked: %v", r)
		}
	}()
	PrintUsage()
}
