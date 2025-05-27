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
└── amd64-windows/
    └── rg.exe
```

Note: Windows ARM64 uses the same amd64-windows binary via emulation, so no separate arm64-windows directory is needed.

## Download Script

The `download.sh` script automatically downloads the latest ripgrep binaries for all supported platforms from the official GitHub releases.

## License

ripgrep is licensed under the MIT License. See `LICENSE-MIT.txt` for details.
