#!/bin/bash

# Install script for layered-code
# This script downloads and installs layered-code along with ripgrep

set -e

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="x86_64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Map OS names for releases
case "$OS" in
    linux)
        OS_NAME="Linux"
        ;;
    darwin)
        OS_NAME="Darwin"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "Installing layered-code for $OS_NAME $ARCH..."

# Check if ripgrep is installed
if ! command -v rg &> /dev/null; then
    echo "ripgrep is required but not installed."
    echo ""
    echo "Please install ripgrep first:"
    echo "  macOS:        brew install ripgrep"
    echo "  Ubuntu/Debian: sudo apt install ripgrep"
    echo "  Fedora:       sudo dnf install ripgrep"
    echo "  Arch:         sudo pacman -S ripgrep"
    echo ""
    echo "Or visit: https://github.com/BurntSushi/ripgrep#installation"
    exit 1
fi

# Get latest release
LATEST_RELEASE=$(curl -s https://api.github.com/repos/layered-flow/layered-code/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to get latest release"
    exit 1
fi

# Download URL
DOWNLOAD_URL="https://github.com/layered-flow/layered-code/releases/download/${LATEST_RELEASE}/layered-code_${OS_NAME}_${ARCH}.tar.gz"

echo "Downloading layered-code ${LATEST_RELEASE}..."

# Download and extract
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if ! curl -L -o layered-code.tar.gz "$DOWNLOAD_URL"; then
    echo "Failed to download layered-code"
    rm -rf "$TMP_DIR"
    exit 1
fi

tar -xzf layered-code.tar.gz

# Install to /usr/local/bin or current directory
echo "Installing layered-code..."

if [ -w /usr/local/bin ] || sudo -n true 2>/dev/null; then
    # Can install to system location
    echo "Installing to /usr/local/bin/layered-code..."
    if [ -w /usr/local/bin ]; then
        mv layered-code /usr/local/bin/
    else
        sudo mv layered-code /usr/local/bin/
    fi
    # Clean up
    cd - > /dev/null
    rm -rf "$TMP_DIR"
else
    # Install to current directory
    echo "Installing to current directory..."
    mv layered-code "$OLDPWD/" 2>/dev/null || mv layered-code ./
    echo "Binary installed as: ./layered-code"
    echo "To install system-wide later, run: sudo mv layered-code /usr/local/bin/"
    # Clean up
    cd - > /dev/null
    rm -rf "$TMP_DIR"
fi

# Verify installation
if layered-code --version &> /dev/null; then
    echo "✅ layered-code installed successfully!"
    layered-code --version
else
    echo "❌ Installation failed"
    exit 1
fi