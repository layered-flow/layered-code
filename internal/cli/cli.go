package cli

import (
	"fmt"
	"os"

	"github.com/layered-flow/layered-code/internal/tools"
)

// PrintUsage displays the available commands and their usage information
func PrintUsage() {
	fmt.Println("Usage: layered-code <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  mcp_server                Start the MCP server")
	fmt.Println("  tool list_apps            List all available apps")
	fmt.Println("  tool list_files           List files and directories within an app")
	fmt.Println("  tool read_file            Read the contents of a file within an app")
	fmt.Println("  help, -h, --help          Show this help message")
	fmt.Println("  version, -v, --version    Show version information")
}

// RunTool executes the specified tool subcommand with the provided arguments
func RunTool() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("tool subcommand is required\nUsage: layered-code tool <subcommand> [args]\nAvailable tools: list_apps, list_files, read_file")
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "list_apps":
		return tools.ListAppsCli()
	case "list_files":
		return tools.ListFilesCli()
	case "read_file":
		return tools.ReadFileCli()
	default:
		return fmt.Errorf("unknown tool: %s\nAvailable tools: list_apps, list_files, read_file", subcommand)
	}
}
