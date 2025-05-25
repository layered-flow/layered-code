# layered-code

[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](LICENSE.md)
[![Go Version](https://img.shields.io/badge/Go-1.24+-brightgreen.svg)](https://golang.org/dl/)
[![Build Status](https://github.com/layered-flow/layered-code/actions/workflows/tests.yml/badge.svg)](https://github.com/layered-flow/layered-code/actions)

> ⚠️ **Work in Progress** ⚠️
>
> This project is currently under active development. Features, APIs, and documentation may change frequently. While we welcome feedback and contributions (see [Contributing](#contributing)), please note that this software is not yet ready for production use. This banner will be removed once we reach version 1.0.

## 🔍 What is layered-code?

Traditional development with AI can feel chaotic - losing context, making large unfocused changes, or struggling to track what actually worked. **layered-code** solves this by transforming how you build and evolve software.

**layered-code** is a CLI tool and MCP server that organizes your codebase into structured, traceable layers. It provides an iterative approach to collaborate with AI while maintaining full control over your development process.

Each layer represents a decision point where you can review, refine, or pivot based on your project's evolving needs.

- **🤝 Pair with AI naturally** through an intuitive CLI interface and Model Context Protocol (MCP)
- **📝 Build incrementally** by creating small, focused changes that compound into complete features
- **🔍 Maintain full context** with a structured change history that AI can understand and build upon
- **🔄 Iterate confidently** as each change preserves working code while evolving functionality
- **📈 Scale seamlessly** from quick prototypes to complex systems with consistent AI collaboration

Think of it as version control for features rather than just files - each layer represents a complete, working increment of functionality that builds upon previous layers while maintaining full traceability of how your software evolved.

## 🎨 Example Usage

**Example:** You could ask: "Add a hero section to the top of the home page with the image 'mountain.png' and the tagline 'Breathtaking views' that links to the mountains page"

The tool would:
1. Create a new layer containing the required steps
2. Execute and track steps in YAML files (moving assets, creating hero section)
3. Allow preview and refinement of the layer's changes
4. Once finalized, commit the layer to git with its summary preserved
5. Make layer history available to future changes as context

The ability to access this historical context enables the AI to understand all prior decisions and actions, resulting in more precise and contextually-aware future changes with fewer errors. Each layer builds upon this knowledge base, creating an evolving understanding of your codebase that influences future changes.

## 🎯 Motivation

**layered-code** provides an open approach that gives you the freedom to choose your own tools, hosting providers, and development workflows:
Your code, your tools, your choice.

- **🔓 Forever Free & Open Source**: layered-code will always remain completely free and open source
- **🛠️ Technology Agnostic**: Works with any language, framework, or development environment
- **🤝 Human-AI Partnership**: Designed for seamless collaboration between developers and AI agents
- **📊 Full Traceability**: Track feature evolution and maintain contextual awareness across your entire codebase
- **🚀 Zero Vendor Lock-in**: Use any hosting provider, development environment, or toolchain

## 📋 Requirements

### System Requirements
- **Operating System**: Linux, macOS or Windows
- **Architecture**: x86_64 or ARM64

### AI Integration Requirements
- **Claude Desktop**: Latest version for MCP integration
- **Model Context Protocol**: Automatically handled by Claude Desktop

### Runtime Dependencies
- **Go**: Version 1.24 or higher (only required for building from source)

## ✨ Quick Start with Claude Desktop

While there are plans to support open source models through Ollama, Claude Desktop currently provides the best support for the Model Context Protocol that **layered-code** relies on.

1. **Install [Claude Desktop](https://claude.ai/download)**

2. **Enable Developer Mode**:
   - Open Claude Desktop
   - Go to Settings
   - Navigate to the "Developer" tab
   - Enable "Developer Mode" if not already enabled

3. **Download layered-code binary**:
   - Visit the [GitHub releases page](https://github.com/layered-flow/layered-code/releases)
   - Download the appropriate binary for your system:
     - **macOS (Intel)**: `layered-code-darwin-amd64`
     - **macOS (Apple Silicon)**: `layered-code-darwin-arm64`
     - **Linux (x86_64)**: `layered-code-linux-amd64`
     - **Linux (ARM64)**: `layered-code-linux-arm64`
     - **Windows (x86_64)**: `layered-code-windows-amd64.exe`
     - **Windows (ARM64)**: `layered-code-windows-arm64.exe`
   - Make the binary executable (macOS/Linux): `chmod +x layered-code-*`
   - Move to a convenient location and remember where you put it (e.g., `~/bin/layered-code` or `/usr/local/bin/layered-code`)

4. **Configure MCP Server**:
   - Open Claude Desktop settings → Developer → Edit Config
   - Add to `claude_desktop_config.json`:

```json
{
  "globalShortcut": "",
  "mcpServers": {
    "layered-code": {
      "command": "/Users/person/layered-code/layered-code",
      "args": ["mcp_server"],
      "env": {
        "LAYERED_APPS_DIRECTORY": "~/LayeredApps"
      }
    }
  }
}
```

*NB. `env` is entirely optional and can be omitted if using the default ~/LayeredApps directory*

5. **Update the path** to your `layered-code` binary location (use double backslashes on Windows, e.g. `C:\\Users\\person\\layered-code\\layered-code.exe`)
6. **Optional**: Set your apps directory in the `env` section (defaults to `~/LayeredApps` if not specified)
7. **Restart Claude Desktop** completely (Windows may require you to "end task" on any Claude background tasks)
8. **Verify**: Check for "layered-code" in Claude's tools menu

### 📁 App Directory Configuration

**layered-code** allows configuring the apps directory while maintaining security:

- Set `LAYERED_APPS_DIRECTORY` environment variable (default: `~/LayeredApps`)
- For security, paths are validated to ensure they're within the user's home directory
- Relative paths are allowed and resolved relative to the user's home directory

### 🖥️ CLI Usage

Use layered-code directly from the command line:

```bash
# Start MCP server
./layered-code mcp_server

# List apps
./layered-code tool list_apps

# List files in an app
./layered-code tool list_files --app-name myapp

# List files with all metadata
./layered-code tool list_files --app-name myapp --include-mime-types --include-size --include-last-modified --include-child-count

# List files matching a pattern
./layered-code tool list_files --app-name myapp --pattern '*.js'

# List files in specific subdirectory using glob pattern
./layered-code tool list_files --app-name myapp --pattern 'src/*.go'

# List all test files recursively
./layered-code tool list_files --app-name myapp --pattern '**/*.test.js'

# Note: list_files automatically skips hidden files/folders and symlinks
# Maximum depth is limited to 10,000 levels for safety

# Get version information
./layered-code version
./layered-code -v
./layered-code --version

# Get help and usage information
./layered-code help
./layered-code -h
./layered-code --help
```

**Available Commands:**
- `mcp_server` - Start the Model Context Protocol server for Claude Desktop integration
- `tool` - Run various tools and utilities (use with subcommands like `list_apps`, `list_files`)
  - `tool list_apps` - List all available applications in the apps directory
  - `tool list_files` - List files and directories within an application with optional metadata (max depth: 10,000 levels)
- `version`, `-v`, `--version` - Display the current version of layered-code
- `help`, `-h`, `--help` - Show usage information and available commands

## 🔨 Building from Source

**Prerequisites:** Go 1.24+

```bash
# Clone and build
git clone https://github.com/layered-flow/layered-code.git
cd layered-code
go build -o layered-code ./cmd/layered_code

# Verify installation
./layered-code
```

## 🧪 Running Tests

Run the test suite to ensure everything is working correctly:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...
```

<a id="contributing"></a>

## 🤝 Contributing

We're excited about the potential for community contributions! While **layered-code** is currently maturing toward version 1.0, we're focusing on stabilizing core functionality and APIs. During this phase, we welcome your ideas and feedback through GitHub Issues, but won't be accepting pull requests to ensure we can move quickly and make necessary breaking changes.

Once we reach version 1.0 and the foundation is solid, we'll be thrilled to open up to community pull requests as well. We believe this focused approach will create a better experience for everyone in the long run.

Feel free to:
- Open issues to share ideas and suggestions
- Report bugs you encounter
- Ask questions about the project
- Provide feedback on features and documentation

Thank you for your understanding and interest in the project! 🙏

## 📝 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

Copyright (c) 2025 Layered Flow<br />
https://www.layeredflow.ai/