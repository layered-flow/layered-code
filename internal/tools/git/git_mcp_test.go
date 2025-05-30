package git

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGitStatusMcp(t *testing.T) {
	ctx := context.Background()
	
	// Missing app_name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}
	_, err := GitStatusMcp(ctx, req)
	if err == nil {
		t.Error("Expected error for missing app_name")
	}
}

func TestGitDiffMcp(t *testing.T) {
	ctx := context.Background()
	
	// Missing app_name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"staged": true,
	}
	_, err := GitDiffMcp(ctx, req)
	if err == nil {
		t.Error("Expected error for missing app_name")
	}
}

func TestGitCommitMcp(t *testing.T) {
	ctx := context.Background()
	
	// Missing app_name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"message": "test",
	}
	_, err := GitCommitMcp(ctx, req)
	if err == nil {
		t.Error("Expected error for missing app_name")
	}
}

func TestGitLogMcp(t *testing.T) {
	ctx := context.Background()
	
	// Missing app_name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}
	_, err := GitLogMcp(ctx, req)
	if err == nil {
		t.Error("Expected error for missing app_name")
	}
}

func TestGitAddMcp(t *testing.T) {
	ctx := context.Background()
	
	// Missing app_name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"files": []string{"test.txt"},
	}
	_, err := GitAddMcp(ctx, req)
	if err == nil {
		t.Error("Expected error for missing app_name")
	}
}

func TestGitInitMcp(t *testing.T) {
	ctx := context.Background()
	
	// Missing app_name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"bare": false,
	}
	_, err := GitInitMcp(ctx, req)
	if err == nil {
		t.Error("Expected error for missing app_name")
	}
}

func TestGitShowMcp(t *testing.T) {
	ctx := context.Background()
	
	// Missing app_name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"commit_ref": "HEAD",
	}
	_, err := GitShowMcp(ctx, req)
	if err == nil {
		t.Error("Expected error for missing app_name")
	}
}