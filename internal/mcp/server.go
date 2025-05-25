package mcp

import (
	"fmt"

	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/layered-flow/layered-code/internal/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StartServer creates and starts the MCP server with all registered tools
func StartServer(name, version string) error {
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
		mcp.WithBoolean("include_mime_types", mcp.Description("Include MIME types for files and directory indicators")),
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
