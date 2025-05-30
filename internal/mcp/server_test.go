package mcp

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TestRegisterFunctions tests that tool registration functions execute without panicking
func TestRegisterFunctions(t *testing.T) {
	s := server.NewMCPServer("test", "1.0.0")

	tests := []struct {
		name string
		fn   func(*server.MCPServer)
	}{
		{"registerListAppsTool", registerListAppsTool},
		{"registerListFilesTool", registerListFilesTool},
		{"registerSearchTextTool", registerSearchTextTool},
		{"registerReadFileTool", registerReadFileTool},
		{"registerWriteFileTool", registerWriteFileTool},
		{"registerEditFileTool", registerEditFileTool},
		{"registerGitStatusTool", registerGitStatusTool},
		{"registerGitDiffTool", registerGitDiffTool},
		{"registerGitCommitTool", registerGitCommitTool},
		{"registerGitLogTool", registerGitLogTool},
		{"registerGitBranchTool", registerGitBranchTool},
		{"registerGitAddTool", registerGitAddTool},
		{"registerGitRestoreTool", registerGitRestoreTool},
		{"registerGitStashTool", registerGitStashTool},
		{"registerGitPushTool", registerGitPushTool},
		{"registerGitPullTool", registerGitPullTool},
		{"registerGitInitTool", registerGitInitTool},
		{"registerGitRemoteTool", registerGitRemoteTool},
		{"registerGitResetTool", registerGitResetTool},
		{"registerGitRevertTool", registerGitRevertTool},
		{"registerGitCheckoutTool", registerGitCheckoutTool},
		{"registerGitShowTool", registerGitShowTool},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn(s)
		})
	}
}

// TestToolCreation verifies tool creation with expected properties
func TestToolCreation(t *testing.T) {
	tool := mcp.NewTool("list_apps",
		mcp.WithDescription("List all available applications"),
	)

	if tool.Name != "list_apps" {
		t.Errorf("Expected tool name 'list_apps', got '%s'", tool.Name)
	}

	if tool.Description != "List all available applications" {
		t.Errorf("Expected description 'List all available applications', got '%s'", tool.Description)
	}
}
