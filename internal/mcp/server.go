package mcp

import (
	"fmt"
	"net/http"

	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/layered-flow/layered-code/internal/notifications"
	"github.com/layered-flow/layered-code/internal/tools"
	"github.com/layered-flow/layered-code/internal/websocket"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var wsHub *websocket.Hub

// StartServer creates and starts the MCP server with all registered tools
func StartServer(name, version string) error {
	// Start WebSocket server for file change notifications
	wsHub = websocket.NewHub()
	go wsHub.Run()
	
	// Set the hub for notifications
	notifications.SetHub(wsHub)
	
	// Start HTTP server for WebSocket connections
	go func() {
		http.HandleFunc("/ws", wsHub.ServeWS)
		http.ListenAndServe(":8080", nil)
	}()

	// Create a new MCP server
	s := server.NewMCPServer(
		name,
		version,
		server.WithToolCapabilities(false),
	)

	// Register all tools
	registerTools(s)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// registerTools registers all available tools with the MCP server
func registerTools(s *server.MCPServer) {
	registerListAppsTool(s)
	registerListFilesTool(s)
	registerReadFileTool(s)
	registerWriteFileTool(s)
	registerEditFileTool(s)
	registerSearchTextTool(s)
}

// registerListAppsTool registers the list_apps tool
func registerListAppsTool(s *server.MCPServer) {
	tool := mcp.NewTool("list_apps",
		mcp.WithDescription("List all available applications"),
	)

	s.AddTool(tool, tools.ListAppsMcp)
}

// registerListFilesTool registers the list_files tool
func registerListFilesTool(s *server.MCPServer) {
	tool := mcp.NewTool("list_files",
		mcp.WithDescription("List files and directories within an application (max depth: 10,000 levels)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("pattern", mcp.Description("Glob pattern to filter files (e.g. '*.txt', 'src/*.js', '**/*.test.js')")),
		mcp.WithBoolean("include_last_modified", mcp.Description("Include last modification timestamps")),
		mcp.WithBoolean("include_size", mcp.Description("Include file and directory sizes")),
		mcp.WithBoolean("include_child_count", mcp.Description("Include count of immediate children for each entry")),
	)

	s.AddTool(tool, tools.ListFilesMcp)
}

// registerReadFileTool registers the read_file tool
func registerReadFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("read_file",
		mcp.WithDescription("Read the contents of a file within an application directory"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file relative to the app directory (must be a text file, cannot be a symlink or binary file, max size "+constants.MaxFileSizeInWords+")")),
	)

	s.AddTool(tool, tools.ReadFileMcp)
}

// registerWriteFileTool registers the write_file tool
func registerWriteFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("write_file",
		mcp.WithDescription("Write or create a file within an application directory"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file relative to the app directory")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content to write to the file (max size "+constants.MaxFileSizeInWords+")")),
		mcp.WithString("mode", mcp.Description("Write mode: 'create' (default, fails if file exists) or 'overwrite' (replaces existing file)")),
	)

	s.AddTool(tool, tools.WriteFileMcp)
}

// registerEditFileTool registers the edit_file tool
func registerEditFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("edit_file",
		mcp.WithDescription("Edit a file by performing find-and-replace operations"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file relative to the app directory")),
		mcp.WithString("old_string", mcp.Required(), mcp.Description("Text to find and replace")),
		mcp.WithString("new_string", mcp.Required(), mcp.Description("Text to replace with (can be empty for deletion)")),
		mcp.WithNumber("occurrences", mcp.Description("Number of occurrences to replace (0 = all, default: 0)")),
	)

	s.AddTool(tool, tools.EditFileMcp)
}

// registerSearchTextTool registers the search_text tool
func registerSearchTextTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_text",
		mcp.WithDescription("Search for text patterns in files within an application directory using ripgrep"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Search pattern (supports regular expressions)")),
		mcp.WithBoolean("case_sensitive", mcp.Description("Perform case-sensitive search (default: false)")),
		mcp.WithBoolean("whole_word", mcp.Description("Match whole words only")),
		mcp.WithString("file_pattern", mcp.Description("Only search files matching this glob pattern (e.g. '*.go', '*.js')")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (default: 100)")),
		mcp.WithBoolean("include_hidden", mcp.Description("Include hidden files and directories in search")),
	)

	s.AddTool(tool, tools.SearchTextMcp)
}
