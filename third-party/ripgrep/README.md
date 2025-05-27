# Ripgrep Binary Distribution

This directory contains pre-compiled ripgrep binaries for different platforms.

## Directory Structure

The binaries should be organized in subdirectories by architecture and OS:

```
third-party/ripgrep/
├── arm64-darwin/
│   └── rg
├── amd64-darwin/
│   └── rg
├── arm64-linux/
│   └── rg
├── amd64-linux/
│   └── rg
├── amd64-windows/
│   └── rg.exe
└── arm64-windows/
    └── rg.exe
```

## Binary Sources

Download ripgrep binaries from the official releases:
https://github.com/BurntSushi/ripgrep/releases

## Version

For consistency, use ripgrep version 14.0.0 or later.

## Build Process

When building layered-code:

1. The build process should include the appropriate ripgrep binary for the target platform
2. The binary should be bundled with the executable
3. The search_text tool will look for the binary in the following order:
   - System PATH (if user has ripgrep installed)
   - Relative to the layered-code executable in third-party/ripgrep/{arch}-{os}/
   - Same directory as the layered-code executable

## Platform Mapping

- macOS Intel: amd64-darwin
- macOS Apple Silicon: arm64-darwin
- Linux x64: amd64-linux
- Linux ARM64: arm64-linux
- Windows x64: amd64-windows
- Windows ARM64: arm64-windows

## File Permissions

Ensure the ripgrep binaries have executable permissions (chmod +x) on Unix-like systems.