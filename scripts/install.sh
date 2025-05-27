#!/bin/bash

# Install script for layered-code
# This script downloads and installs layered-code with bundled ripgrep

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

# Determine ripgrep binary name
RG_BINARY="rg"
if [ "$OS" = "windows" ]; then
    RG_BINARY="rg.exe"
fi

# Install to /usr/local/bin or current directory
echo "Installing layered-code and bundled ripgrep..."

if [ -w /usr/local/bin ] || sudo -n true 2>/dev/null; then
    # Can install to system location
    echo "Installing to /usr/local/bin/..."
    if [ -w /usr/local/bin ]; then
        mv layered-code /usr/local/bin/
        # Install ripgrep in the same directory as layered-code for bundled functionality
        if [ -f "$RG_BINARY" ]; then
            mv "$RG_BINARY" /usr/local/bin/
            chmod +x "/usr/local/bin/$RG_BINARY"
        fi
    else
        sudo mv layered-code /usr/local/bin/
        # Install ripgrep in the same directory as layered-code for bundled functionality
        if [ -f "$RG_BINARY" ]; then
            sudo mv "$RG_BINARY" /usr/local/bin/
            sudo chmod +x "/usr/local/bin/$RG_BINARY"
        fi
    fi
    # Clean up
    cd - > /dev/null
    rm -rf "$TMP_DIR"
else
    # Install to current directory
    echo "Installing to current directory..."
    mv layered-code "$OLDPWD/" 2>/dev/null || mv layered-code ./
    if [ -f "$RG_BINARY" ]; then
        mv "$RG_BINARY" "$OLDPWD/" 2>/dev/null || mv "$RG_BINARY" ./
        chmod +x "./$RG_BINARY"
    fi
    echo "Binaries installed as: ./layered-code and ./$RG_BINARY"
    echo "To install system-wide later, run: sudo mv layered-code $RG_BINARY /usr/local/bin/"
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