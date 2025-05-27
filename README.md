[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](LICENSE.md)
[![Go Version](https://img.shields.io/badge/Go-1.24+-brightgreen.svg)](https://golang.org/dl/)
[![Build Status](https://github.com/layered-flow/layered-code/actions/workflows/tests.yml/badge.svg)](https://github.com/layered-flow/layered-code/actions)

![Banner](/docs/images/banner.png)

> ⚠️ **Work in Progress** ⚠️
>
> This project is currently under active development. Features, APIs, and documentation may change frequently. While we welcome feedback and contributions (see [Contributing](#contributing)), please note that this software is not yet ready for production use. This banner will be removed once we reach version 1.0.

## 🔍 What is Layered Code?

Traditional development with AI can feel chaotic - losing context, making large unfocused changes, or struggling to track what actually worked. **Layered Code** solves this by transforming how you build and evolve software through the Model Context Protocol (MCP). This protocol enables AI assistants to maintain deep contextual awareness of your codebase while making focused, traceable changes.

## 💬 How You Interact with Layered Code

**Layered Code** is designed to be used primarily through **conversational AI interfaces** that support the Model Context Protocol (MCP). Rather than memorizing complex commands or navigating intricate UIs, you simply describe what you want to build in natural language.

**Primary interaction method:**
- **🗣️ Natural language conversations** with AI assistants like Claude Desktop
- **🔌 MCP integration** provides the AI with direct access to your codebase and layered history
- **🎯 Intent-driven development** - describe your goals, and the AI handles the technical implementation

**Example conversation:**
> "I want to redesign the homepage with a modern hero section, update the navigation menu, and add a testimonials carousel below the fold"

The AI, equipped with Layered Code's MCP tools, can:
- Understand your current page structure and styling
- Access the history of previous design layers and decisions
- Create a new layer with the homepage redesign
- Handle HTML structure changes, CSS updates, and asset organization
- Preserve your existing content while enhancing the visual design

While a CLI interface is available for direct tool access, the conversational AI experience through MCP is where **Layered Code** truly shines - enabling you to focus on what you want to build rather than how to build it.

## 🎯 Motivation

**Layered Code** provides an open approach that gives you the freedom to choose your own tools, hosting providers, and development workflows: **Your code, your tools, your choice.**

- **🔓 Forever Free & Open Source**: Layered Code will always remain completely free and open source
- **🛠️ Technology Agnostic**: Works with any language, framework, or development environment
- **🤝 Human-AI Partnership**: Designed for seamless collaboration between developers and AI agents
- **📊 Full Traceability**: Track feature evolution and maintain contextual awareness across your entire codebase
- **🚀 Zero Vendor Lock-in**: Use any hosting provider, development environment, or toolchain

## 🏗️ Building from Source

To build Layered Code from source:

```bash
# Clone the repository
git clone https://github.com/layered-flow/layered-code.git
cd layered-code

# Build with make (downloads ripgrep binaries automatically)
make build

# Or build for all platforms
make build-all

# Run tests
make test
```

## 📦 Installation

### Prerequisites

- **Required**: [ripgrep](https://github.com/BurntSushi/ripgrep) for text search functionality
  - Install via your package manager:
    - macOS: `brew install ripgrep`
    - Ubuntu/Debian: `sudo apt install ripgrep`
    - Windows: `choco install ripgrep` or `scoop install ripgrep`
    - Or download from: https://github.com/BurntSushi/ripgrep/releases

### Option 1: macOS/Linux via Homebrew (Recommended)

```bash
brew update
brew tap layered-flow/layered-code
brew install layered-code
```

Note: Homebrew will automatically install ripgrep as a dependency.

### Option 2: Install Script (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/layered-flow/layered-code/main/scripts/install.sh | bash
```

Note: This script will check for ripgrep and guide you to install it if needed.

### Option 3: Pre-built Binaries

Download the appropriate binary for your platform from the [GitHub releases page](https://github.com/layered-flow/layered-code/releases):

- **macOS (Intel)**: `layered-code-darwin-amd64`
- **macOS (Apple Silicon)**: `layered-code-darwin-arm64`
- **Linux (x86_64)**: `layered-code-linux-amd64`
- **Linux (ARM64)**: `layered-code-linux-arm64`
- **Windows (x86_64)**: `layered-code-windows-amd64.exe`
- **Windows (ARM64)**: `layered-code-windows-arm64.exe`

**Setup steps:**

**macOS/Linux:**
1. Make the binary executable: `chmod +x layered-code-*`
2. Move to a convenient location (e.g., `~/bin/layered-code` or `/usr/local/bin/layered-code`)
3. Ensure the location is in your PATH for easy access

**Windows:**
1. Download the `.exe` file (no chmod needed - Windows executables are ready to run)
2. Move to a convenient location (e.g., `C:\Users\YourName\bin\layered-code.exe`)
3. Add the directory to your PATH environment variable:
   - Open "Environment Variables" in System Properties
   - Add the directory containing `layered-code.exe` to your PATH
   - Or run directly using the full path: `C:\path\to\layered-code.exe`

## ✨ Quick Start with Claude Desktop

While there are plans to support open source models through Ollama, Claude Desktop currently provides the best support for the Model Context Protocol that **Layered Code** relies on.

1. **Install [Claude Desktop](https://claude.ai/download)**

2. **Enable Developer Mode**:
   - Open Claude Desktop
   - Go to Settings
   - Navigate to the "Developer" tab
   - Enable "Developer Mode" if not already enabled

3. **Install layered-code**: Follow the [Installation](#-installation) instructions above and note the path to your binary

4. **Configure MCP Server**:
   - Open Claude Desktop settings → Developer → Edit Config
   - Add to `claude_desktop_config.json`:

### Configure MCP Server in Claude Desktop

Add the following to your `claude_desktop_config.json` (under Settings → Developer → Edit Config), adjusting the `command` path as appropriate for your installation and platform:

| Platform/Install Method         | "command" value example                        |
|---------------------------------|--------------------------------------------------|
| macOS/Linux (Homebrew)          | "layered-code"                                 |
| macOS/Linux (Manual/Binary)     | "/usr/local/bin/layered-code"                  |
| Windows                         | "C:\\Users\\person\\bin\\layered-code.exe"     |

> **Note for Windows:** Use the full path with double backslashes (`\\`) in the `"command"` value.

**Example configuration:**
```json
{
  "globalShortcut": "",
  "mcpServers": {
    "layered-code": {
      "command": "<see table above>",
      "args": ["mcp_server"]
    }
  }
}
```

5. **Restart Claude Desktop** completely (Windows may require you to "end task" on any Claude background tasks)
6. **Verify**: Check for "layered-code" in Claude's tools menu

### 🔧 Optional: Custom Apps Directory

By default, **Layered Code** uses `~/LayeredApps` as the directory for your applications. To use a custom directory:

**For MCP (Claude Desktop):** Add an `env` section to your configuration:

```json
{
  "globalShortcut": "",
  "mcpServers": {
    "layered-code": {
      "command": "layered-code",
      "args": ["mcp_server"],
      "env": {
        "LAYERED_APPS_DIRECTORY": "~/MyCustomAppsFolder"
      }
    }
  }
}
```

**For CLI usage:** Set the `LAYERED_APPS_DIRECTORY` environment variable in your shell:

```bash
export LAYERED_APPS_DIRECTORY="~/MyCustomAppsFolder"
layered-code tool list_apps
```

### 🔒 Security

**Layered Code** maintains security when configuring custom app directories:

- For security, paths are validated to ensure they're within the user's home directory
- Relative paths are allowed and resolved relative to the user's home directory

### 🖥️ CLI Usage

Use layered-code directly from the command line:

```bash
# Start MCP server
layered-code mcp_server

# List apps
layered-code tool list_apps

# Get version information
layered-code version
layered-code -v
layered-code --version

# Get help and usage information
layered-code help
layered-code -h
layered-code --help
```

**Available Commands:**
- `mcp_server` - Start the Model Context Protocol server for Claude Desktop integration
- `tool` - Run various tools and utilities (use with subcommands like below)
  - `tool list_apps` - List all available applications in the apps directory
  - `tool list_files` - List files and directories within an application with optional metadata (max depth: 10,000 levels)
  - `tool search_text` - Search for text patterns in files within an application directory using ripgrep
  - `tool read_file` - Read the contents of a file within an application directory
  - `tool write_file` - Write or create a file within an application directory
  - `tool edit_file` - Edit a file by performing find-and-replace operations
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
layered-code
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

We're excited about the potential for community contributions! While **Layered Code** is currently maturing toward version 1.0, we're focusing on stabilizing core functionality and APIs. During this phase, we welcome your ideas and feedback through GitHub Issues, but won't be accepting pull requests to ensure we can move quickly and make necessary breaking changes.

Once we reach version 1.0 and the foundation is solid, we'll be thrilled to open up to community pull requests as well. We believe this focused approach will create a better experience for everyone in the long run.

Feel free to:
- Open issues to share ideas and suggestions
- Report bugs you encounter
- Ask questions about the project
- Provide feedback on features and documentation

Thank you for your understanding and interest in the project! 🙏

## 📝 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

### Third-Party Components

This software includes third-party components. See [THIRD-PARTY-LICENSES.md](THIRD-PARTY-LICENSES.md) for their license terms.

Copyright (c) 2025 Layered Flow<br />
https://www.layeredflow.ai/