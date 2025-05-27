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
	// File management tools
	registerListAppsTool(s)
	registerListFilesTool(s)
	registerSearchTextTool(s)
	registerReadFileTool(s)
	registerWriteFileTool(s)
	registerEditFileTool(s)
	
	// Git tools
	registerGitStatusTool(s)
	registerGitDiffTool(s)
	registerGitCommitTool(s)
	registerGitLogTool(s)
	registerGitBranchTool(s)
	registerGitAddTool(s)
	registerGitRestoreTool(s)
	registerGitStashTool(s)
	registerGitPushTool(s)
	registerGitPullTool(s)
	registerGitInitTool(s)
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

// registerGitStatusTool registers the git_status tool
func registerGitStatusTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_status",
		mcp.WithDescription("Show the working tree status of a git repository (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
	)

	s.AddTool(tool, tools.GitStatusMcp)
}

// registerGitDiffTool registers the git_diff tool
func registerGitDiffTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_diff",
		mcp.WithDescription("Show changes between commits, commit and working tree, etc (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithBoolean("staged", mcp.Description("Show staged changes instead of unstaged")),
		mcp.WithString("file_path", mcp.Description("Specific file to diff (relative to app directory)")),
	)

	s.AddTool(tool, tools.GitDiffMcp)
}

// registerGitCommitTool registers the git_commit tool
func registerGitCommitTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_commit",
		mcp.WithDescription("Create a new commit with staged changes (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("message", mcp.Description("Commit message (required unless using --amend)")),
		mcp.WithBoolean("amend", mcp.Description("Amend the previous commit")),
	)

	s.AddTool(tool, tools.GitCommitMcp)
}

// registerGitLogTool registers the git_log tool
func registerGitLogTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_log",
		mcp.WithDescription("Show commit logs (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of commits to show (default: 10)")),
		mcp.WithBoolean("oneline", mcp.Description("Show commits in one-line format")),
	)

	s.AddTool(tool, tools.GitLogMcp)
}

// registerGitBranchTool registers the git_branch tool
func registerGitBranchTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_branch",
		mcp.WithDescription("List, create, or delete branches (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("create_branch", mcp.Description("Name of branch to create")),
		mcp.WithString("switch_branch", mcp.Description("Name of branch to switch to")),
		mcp.WithString("delete_branch", mcp.Description("Name of branch to delete")),
		mcp.WithBoolean("list_all", mcp.Description("List all branches including remotes")),
	)

	s.AddTool(tool, tools.GitBranchMcp)
}

// registerGitAddTool registers the git_add tool
func registerGitAddTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_add",
		mcp.WithDescription("Add file contents to the staging area (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithObject("files", mcp.Description("List of files to add (relative to app directory) - pass as JSON array")),
		mcp.WithBoolean("all", mcp.Description("Add all changes (equivalent to -A)")),
	)

	s.AddTool(tool, tools.GitAddMcp)
}

// registerGitRestoreTool registers the git_restore tool
func registerGitRestoreTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_restore",
		mcp.WithDescription("Restore working tree files (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithObject("files", mcp.Required(), mcp.Description("List of files to restore (relative to app directory) - pass as JSON array")),
		mcp.WithBoolean("staged", mcp.Description("Restore files in the staging area")),
	)

	s.AddTool(tool, tools.GitRestoreMcp)
}

// registerGitStashTool registers the git_stash tool
func registerGitStashTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_stash",
		mcp.WithDescription("Stash the changes in a dirty working directory (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("action", mcp.Description("Action to perform: push, pop, apply, drop, list (default: list)")),
		mcp.WithString("message", mcp.Description("Stash message (for push action)")),
	)

	s.AddTool(tool, tools.GitStashMcp)
}

// registerGitPushTool registers the git_push tool
func registerGitPushTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_push",
		mcp.WithDescription("Update remote refs along with associated objects (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("remote", mcp.Description("Remote name (default: origin)")),
		mcp.WithString("branch", mcp.Description("Branch name to push")),
		mcp.WithBoolean("set_upstream", mcp.Description("Set upstream tracking branch")),
		mcp.WithBoolean("force", mcp.Description("Force push (use with caution)")),
	)

	s.AddTool(tool, tools.GitPushMcp)
}

// registerGitPullTool registers the git_pull tool
func registerGitPullTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_pull",
		mcp.WithDescription("Fetch from and integrate with another repository or local branch (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from list_apps)")),
		mcp.WithString("remote", mcp.Description("Remote name (default: origin)")),
		mcp.WithString("branch", mcp.Description("Branch name to pull")),
		mcp.WithBoolean("rebase", mcp.Description("Rebase instead of merge")),
	)

	s.AddTool(tool, tools.GitPullMcp)
}

// registerGitInitTool registers the git_init tool
func registerGitInitTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_init",
		mcp.WithDescription("Initialize a new git repository"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory to initialize (will be created if it doesn't exist)")),
		mcp.WithBoolean("bare", mcp.Description("Create a bare repository")),
	)

	s.AddTool(tool, tools.GitInitMcp)
}
