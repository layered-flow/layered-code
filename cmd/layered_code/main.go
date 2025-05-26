package main

import (
	"fmt"
	"os"

	"github.com/layered-flow/layered-code/internal/cli"
	"github.com/layered-flow/layered-code/internal/constants"
	"github.com/layered-flow/layered-code/internal/mcp"
	"github.com/layered-flow/layered-code/internal/update"
)

// run contains the main application logic
func run(args []string) error {
	// Check for updates
	if hasUpdate, latestVersion, err := update.CheckForUpdate(constants.ProjectVersion); err == nil && hasUpdate {
		update.DisplayUpdateWarning(latestVersion)
	}

	// return if no arguments are provided
	if len(args) < 2 {
		cli.PrintUsage()
		return fmt.Errorf("no command provided")
	}

	// run the MCP server or a tool with a subcommand
	switch args[1] {
	case "mcp_server":
		if err := mcp.StartServer(constants.ProjectName, constants.ProjectVersion); err != nil {
			return fmt.Errorf("mcp server error: %w", err)
		}
	case "tool":
		if err := cli.RunTool(); err != nil {
			return fmt.Errorf("tool error: %w", err)
		}
	case "help", "-h", "--help":
		cli.PrintUsage()
	case "version", "-v", "--version":
		fmt.Printf("%s version %s\n", constants.ProjectName, constants.ProjectVersion)
	default:
		fmt.Printf("Unknown command: %s\n", args[1])
		cli.PrintUsage()
		return fmt.Errorf("unknown command: %s", args[1])
	}
	return nil
}

// main is the entry point
func main() {
	if err := run(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
