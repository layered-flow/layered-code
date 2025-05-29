package cli

import (
	"fmt"
	"os"

	"github.com/layered-flow/layered-code/internal/tools"
	"github.com/layered-flow/layered-code/internal/tools/git"
)

// PrintUsage displays the available commands and their usage information
func PrintUsage() {
	fmt.Println("Usage: layered-code <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  mcp_server                Start the MCP server")
	fmt.Println()
	fmt.Println("  File Management Tools:")
	fmt.Println("  tool list_apps            List all available apps")
	fmt.Println("  tool list_files           List files and directories within an app")
	fmt.Println("  tool search_text          Search for text patterns in files using ripgrep")
	fmt.Println("  tool read_file            Read the contents of a file within an app")
	fmt.Println("  tool write_file           Write or create a file within an app")
	fmt.Println("  tool edit_file            Edit a file using find-and-replace")
	fmt.Println()
	fmt.Println("  Git Tools:")
	fmt.Println("  tool git_status           Show the working tree status")
	fmt.Println("  tool git_diff             Show changes between commits")
	fmt.Println("  tool git_commit           Create a new commit")
	fmt.Println("  tool git_log              Show commit logs")
	fmt.Println("  tool git_branch           List, create, or delete branches")
	fmt.Println("  tool git_add              Add file contents to staging area")
	fmt.Println("  tool git_restore          Restore working tree files")
	fmt.Println("  tool git_stash            Stash changes in working directory")
	fmt.Println("  tool git_push             Update remote refs")
	fmt.Println("  tool git_pull             Fetch from and integrate with remote")
	fmt.Println("  tool git_init             Initialize a new git repository")
	fmt.Println("  tool git_remote           Manage git remotes (list, add, remove, rename)")
	fmt.Println()
	fmt.Println("  help, -h, --help          Show this help message")
	fmt.Println("  version, -v, --version    Show version information")
}

// RunTool executes the specified tool subcommand with the provided arguments
func RunTool() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("tool subcommand is required\nUsage: layered-code tool <subcommand> [args]\nRun 'layered-code help' to see all available tools")
	}

	subcommand := os.Args[2]
	switch subcommand {
	// File management tools
	case "list_apps":
		return tools.ListAppsCli()
	case "list_files":
		return tools.ListFilesCli()
	case "search_text":
		return tools.SearchTextCli()
	case "read_file":
		return tools.ReadFileCli()
	case "write_file":
		return tools.WriteFileCli()
	case "edit_file":
		return tools.EditFileCli()
	
	// Git tools
	case "git_status":
		return git.GitStatusCli()
	case "git_diff":
		return git.GitDiffCli()
	case "git_commit":
		return git.GitCommitCli()
	case "git_log":
		return git.GitLogCli()
	case "git_branch":
		return git.GitBranchCli()
	case "git_add":
		return git.GitAddCli()
	case "git_restore":
		return git.GitRestoreCli()
	case "git_stash":
		return git.GitStashCli()
	case "git_push":
		return git.GitPushCli()
	case "git_pull":
		return git.GitPullCli()
	case "git_init":
		return git.GitInitCli()
	case "git_remote":
		return git.GitRemoteCli()
		
	default:
		return fmt.Errorf("unknown tool: %s\nRun 'layered-code help' to see all available tools", subcommand)
	}
}
