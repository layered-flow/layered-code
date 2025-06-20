package mcp

import (
	"fmt"
	"net/http"

	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/layered-flow/layered-code/internal/notifications"
	"github.com/layered-flow/layered-code/internal/tools/git"
	"github.com/layered-flow/layered-code/internal/tools/lc"
	"github.com/layered-flow/layered-code/internal/tools/pnpm"
	"github.com/layered-flow/layered-code/internal/tools/vite"
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
	registerMoveFileTool(s)
	registerDeleteFileTool(s)
	registerCopyFileTool(s)
	
	// Vite tools
	registerViteCreateAppTool(s)
	
	// Package Manager tools
	registerPnpmInstallTool(s)
	registerPnpmAddTool(s)
	registerPnpmPm2Tool(s)
	
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
	registerGitRemoteTool(s)
	registerGitResetTool(s)
	registerGitRevertTool(s)
	registerGitCheckoutTool(s)
	registerGitShowTool(s)
}

// registerListAppsTool registers the lc_list_apps tool
func registerListAppsTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_list_apps",
		mcp.WithDescription("List all available applications"),
	)

	s.AddTool(tool, lc.LcListAppsMcp)
}

// registerListFilesTool registers the lc_list_files tool
func registerListFilesTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_list_files",
		mcp.WithDescription("List files and directories within an application (max depth: 10,000 levels)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("pattern", mcp.Description("Glob pattern to filter files (e.g. '*.txt', 'src/*.js', '**/*.test.js')")),
		mcp.WithBoolean("include_last_modified", mcp.Description("Include last modification timestamps")),
		mcp.WithBoolean("include_size", mcp.Description("Include file and directory sizes")),
		mcp.WithBoolean("include_child_count", mcp.Description("Include count of immediate children for each entry")),
	)

	s.AddTool(tool, lc.LcListFilesMcp)
}

// registerSearchTextTool registers the lc_search_text tool
func registerSearchTextTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_search_text",
		mcp.WithDescription("Search for text patterns in files within an application directory using ripgrep"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Search pattern (supports regular expressions)")),
		mcp.WithBoolean("case_sensitive", mcp.Description("Perform case-sensitive search (default: false)")),
		mcp.WithBoolean("whole_word", mcp.Description("Match whole words only")),
		mcp.WithString("file_pattern", mcp.Description("Only search files matching this glob pattern (e.g. '*.go', '*.js')")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (default: 100)")),
		mcp.WithBoolean("include_hidden", mcp.Description("Include hidden files and directories in search")),
	)

	s.AddTool(tool, lc.LcSearchTextMcp)
}

// registerReadFileTool registers the lc_read_file tool
func registerReadFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_read_file",
		mcp.WithDescription("Read the contents of a file within an application directory"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file relative to the app directory (must be a text file, cannot be a symlink or binary file, max size "+constants.MaxFileSizeInWords+")")),
	)

	s.AddTool(tool, lc.LcReadFileMcp)
}

// registerWriteFileTool registers the lc_write_file tool
func registerWriteFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_write_file",
		mcp.WithDescription("Write or create a file within an application directory"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file relative to the app directory")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Content to write to the file (max size "+constants.MaxFileSizeInWords+")")),
		mcp.WithString("mode", mcp.Description("Write mode: 'create' (default, fails if file exists) or 'overwrite' (replaces existing file)")),
	)

	s.AddTool(tool, lc.LcWriteFileMcp)
}

// registerEditFileTool registers the lc_edit_file tool
func registerEditFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_edit_file",
		mcp.WithDescription("Edit a file by performing find-and-replace operations"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file relative to the app directory")),
		mcp.WithString("old_string", mcp.Required(), mcp.Description("Text to find and replace")),
		mcp.WithString("new_string", mcp.Required(), mcp.Description("Text to replace with (can be empty for deletion)")),
		mcp.WithNumber("occurrences", mcp.Description("Number of occurrences to replace (0 = all, default: 0)")),
	)

	s.AddTool(tool, lc.LcEditFileMcp)
}

// registerMoveFileTool registers the lc_move_file tool
func registerMoveFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_move_file",
		mcp.WithDescription("Move or rename a file within an application directory"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("source_path", mcp.Required(), mcp.Description("Source file path relative to the app directory")),
		mcp.WithString("dest_path", mcp.Required(), mcp.Description("Destination file path relative to the app directory")),
	)

	s.AddTool(tool, lc.LcMoveFileMcp)
}

// registerDeleteFileTool registers the lc_delete_file tool
func registerDeleteFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_delete_file",
		mcp.WithDescription("Delete a file within an application directory"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the file relative to the app directory")),
	)

	s.AddTool(tool, lc.LcDeleteFileMcp)
}

// registerCopyFileTool registers the lc_copy_file tool
func registerCopyFileTool(s *server.MCPServer) {
	tool := mcp.NewTool("lc_copy_file",
		mcp.WithDescription("Copy a file within an application directory"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("source_path", mcp.Required(), mcp.Description("Source file path relative to the app directory")),
		mcp.WithString("dest_path", mcp.Required(), mcp.Description("Destination file path relative to the app directory")),
		mcp.WithBoolean("overwrite", mcp.Description("Overwrite destination if it exists (default: false)")),
	)

	s.AddTool(tool, lc.LcCopyFileMcp)
}

// registerGitStatusTool registers the git_status tool
func registerGitStatusTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_status",
		mcp.WithDescription("Show the working tree status of a git repository (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
	)

	s.AddTool(tool, git.GitStatusMcp)
}

// registerGitDiffTool registers the git_diff tool
func registerGitDiffTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_diff",
		mcp.WithDescription("Show changes between commits, commit and working tree, etc (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithBoolean("staged", mcp.Description("Show staged changes instead of unstaged")),
		mcp.WithString("file_path", mcp.Description("Specific file to diff (relative to app directory)")),
	)

	s.AddTool(tool, git.GitDiffMcp)
}

// registerGitCommitTool registers the git_commit tool
func registerGitCommitTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_commit",
		mcp.WithDescription("Create a new commit with staged changes (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("message", mcp.Description("Commit message (required unless using --amend)")),
		mcp.WithBoolean("amend", mcp.Description("Amend the previous commit")),
	)

	s.AddTool(tool, git.GitCommitMcp)
}

// registerGitLogTool registers the git_log tool
func registerGitLogTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_log",
		mcp.WithDescription("Show commit logs (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of commits to show (default: 10)")),
		mcp.WithBoolean("oneline", mcp.Description("Show commits in one-line format")),
	)

	s.AddTool(tool, git.GitLogMcp)
}

// registerGitBranchTool registers the git_branch tool
func registerGitBranchTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_branch",
		mcp.WithDescription("List, create, or delete branches (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("create_branch", mcp.Description("Name of branch to create")),
		mcp.WithString("switch_branch", mcp.Description("Name of branch to switch to")),
		mcp.WithString("delete_branch", mcp.Description("Name of branch to delete")),
		mcp.WithBoolean("list_all", mcp.Description("List all branches including remotes")),
	)

	s.AddTool(tool, git.GitBranchMcp)
}

// registerGitAddTool registers the git_add tool
func registerGitAddTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_add",
		mcp.WithDescription("Add file contents to the staging area (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithObject("files", mcp.Description("List of files to add (relative to app directory) - pass as JSON array")),
		mcp.WithBoolean("all", mcp.Description("Add all changes (equivalent to -A)")),
	)

	s.AddTool(tool, git.GitAddMcp)
}

// registerGitRestoreTool registers the git_restore tool
func registerGitRestoreTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_restore",
		mcp.WithDescription("Restore working tree files (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithObject("files", mcp.Required(), mcp.Description("List of files to restore (relative to app directory) - pass as JSON array")),
		mcp.WithBoolean("staged", mcp.Description("Restore files in the staging area")),
	)

	s.AddTool(tool, git.GitRestoreMcp)
}

// registerGitStashTool registers the git_stash tool
func registerGitStashTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_stash",
		mcp.WithDescription("Stash the changes in a dirty working directory (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("action", mcp.Description("Action to perform: push, pop, apply, drop, list (default: list)")),
		mcp.WithString("message", mcp.Description("Stash message (for push action)")),
	)

	s.AddTool(tool, git.GitStashMcp)
}

// registerGitPushTool registers the git_push tool
func registerGitPushTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_push",
		mcp.WithDescription("Update remote refs along with associated objects (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("remote", mcp.Description("Remote name (default: origin)")),
		mcp.WithString("branch", mcp.Description("Branch name to push")),
		mcp.WithBoolean("set_upstream", mcp.Description("Set upstream tracking branch")),
		mcp.WithBoolean("force", mcp.Description("Force push (use with caution)")),
	)

	s.AddTool(tool, git.GitPushMcp)
}

// registerGitPullTool registers the git_pull tool
func registerGitPullTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_pull",
		mcp.WithDescription("Fetch from and integrate with another repository or local branch (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("remote", mcp.Description("Remote name (default: origin)")),
		mcp.WithString("branch", mcp.Description("Branch name to pull")),
		mcp.WithBoolean("rebase", mcp.Description("Rebase instead of merge")),
	)

	s.AddTool(tool, git.GitPullMcp)
}

// registerGitInitTool registers the git_init tool
func registerGitInitTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_init",
		mcp.WithDescription("Initialize a new git repository"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory to initialize (will be created if it doesn't exist)")),
		mcp.WithBoolean("bare", mcp.Description("Create a bare repository")),
	)

	s.AddTool(tool, git.GitInitMcp)
}

// registerGitRemoteTool registers the git_remote tool
func registerGitRemoteTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_remote",
		mcp.WithDescription("Manage git remotes (list, add, remove, rename, set-url) (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("add_name", mcp.Description("Name of remote to add")),
		mcp.WithString("add_url", mcp.Description("URL of remote to add (required with add_name)")),
		mcp.WithString("remove_name", mcp.Description("Name of remote to remove")),
		mcp.WithString("old_name", mcp.Description("Current name of remote to rename")),
		mcp.WithString("new_name", mcp.Description("New name for remote (required with old_name)")),
		mcp.WithString("set_url_name", mcp.Description("Name of remote to update URL for")),
		mcp.WithString("set_url", mcp.Description("New URL for remote (required with set_url_name)")),
	)

	s.AddTool(tool, git.GitRemoteMcp)
}

// registerGitResetTool registers the git_reset tool
func registerGitResetTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_reset",
		mcp.WithDescription("Reset current HEAD to the specified state (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("commit_hash", mcp.Required(), mcp.Description("Commit hash to reset to")),
		mcp.WithString("mode", mcp.Description("Reset mode: 'soft', 'mixed' (default), or 'hard'")),
	)

	s.AddTool(tool, git.GitResetMcp)
}

// registerGitRevertTool registers the git_revert tool
func registerGitRevertTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_revert",
		mcp.WithDescription("Create a new commit that undoes changes from a previous commit (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("commit_hash", mcp.Required(), mcp.Description("Commit hash to revert")),
		mcp.WithBoolean("no_commit", mcp.Description("Don't create a commit, just stage the changes")),
	)

	s.AddTool(tool, git.GitRevertMcp)
}

// registerGitCheckoutTool registers the git_checkout tool
func registerGitCheckoutTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_checkout",
		mcp.WithDescription("Switch branches or restore working tree files (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("target", mcp.Description("Branch name or commit hash to checkout")),
		mcp.WithBoolean("is_new_branch", mcp.Description("Create a new branch with the given name")),
		mcp.WithObject("files", mcp.Description("List of files to checkout (relative to app directory) - pass as JSON array")),
	)

	s.AddTool(tool, git.GitCheckoutMcp)
}

// registerGitShowTool registers the git_show tool
func registerGitShowTool(s *server.MCPServer) {
	tool := mcp.NewTool("git_show",
		mcp.WithDescription("Show various types of objects (commits, trees, blobs) with their content (requires git to be installed)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("commit_ref", mcp.Description("Commit reference to show (hash, branch, tag, etc.). Defaults to HEAD if not specified")),
	)

	s.AddTool(tool, git.GitShowMcp)
}

// registerViteCreateAppTool registers the vite_create_app tool
func registerViteCreateAppTool(s *server.MCPServer) {
	tool := mcp.NewTool("vite_create_app",
		mcp.WithDescription("Create a new Vite app in the apps directory with specified template (dependencies not installed - use pnpm_install)"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app to create (must be unique, cannot contain special characters or '..')")),
		mcp.WithString("template", mcp.Description("Vite template to use (default: react-ts). Options: vanilla, vanilla-ts, vue, vue-ts, react, react-ts, react-swc, react-swc-ts, preact, preact-ts, lit, lit-ts, svelte, svelte-ts, solid, solid-ts, qwik, qwik-ts")),
	)

	s.AddTool(tool, vite.ViteCreateAppMcp)
}

// registerPnpmInstallTool registers the pnpm_install tool
func registerPnpmInstallTool(s *server.MCPServer) {
	tool := mcp.NewTool("pnpm_install",
		mcp.WithDescription("Install dependencies in an app directory using pnpm (preferred) or npm"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory to install dependencies in (must exactly match an app name from lc_list_apps)")),
	)

	s.AddTool(tool, pnpm.PnpmInstallMcp)
}

// registerPnpmAddTool registers the pnpm_add tool
func registerPnpmAddTool(s *server.MCPServer) {
	tool := mcp.NewTool("pnpm_add",
		mcp.WithDescription("Add a package to an app directory using pnpm (preferred) or npm"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of the app directory to add package to (must exactly match an app name from lc_list_apps)")),
		mcp.WithString("package_name", mcp.Required(), mcp.Description("Name of the package to add (e.g. 'express', 'react@18', '@types/node')")),
	)

	s.AddTool(tool, pnpm.PnpmAddMcp)
}

// registerPnpmPm2Tool registers the pnpm_pm2 tool
func registerPnpmPm2Tool(s *server.MCPServer) {
	tool := mcp.NewTool("pnpm_pm2",
		mcp.WithDescription("Manage Node.js apps with PM2. Smart defaults: 'start' auto-detects dev/start scripts, uses ecosystem.config.js if present"),
		mcp.WithString("command", mcp.Required(), mcp.Description("Command: 'start', 'stop', 'restart', 'delete', 'list', 'logs'")),
		mcp.WithString("target", mcp.Description("App name for start/stop/restart/delete, or 'all' for stop/restart/delete. Not needed for 'list'")),
	)

	s.AddTool(tool, pnpm.PnpmPm2Mcp)
}
