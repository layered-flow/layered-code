#!/bin/bash

# Download ripgrep binaries for all supported platforms
# This script should be run during the build process

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Get the latest release version from GitHub API
echo "Fetching latest ripgrep version from GitHub..."
RIPGREP_VERSION=$(curl -s https://api.github.com/repos/BurntSushi/ripgrep/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$RIPGREP_VERSION" ]; then
    echo "Failed to fetch latest version, falling back to v14.1.0"
    RIPGREP_VERSION="14.1.0"
else
    echo "Latest version found: $RIPGREP_VERSION"
fi

echo "Downloading ripgrep v${RIPGREP_VERSION} binaries..."

# Function to download and extract ripgrep
download_ripgrep() {
    local platform=$1
    local url=$2
    local archive_name=$3
    local binary_path=$4

    echo "Downloading for ${platform}..."

    # Create platform directory
    mkdir -p "${SCRIPT_DIR}/${platform}"

    # Download archive
    curl -L -o "/tmp/${archive_name}" "${url}"

    # Extract binary
    case "${archive_name}" in
        *.tar.gz)
            tar -xzf "/tmp/${archive_name}" -C "/tmp"
            ;;
        *.zip)
            unzip -q "/tmp/${archive_name}" -d "/tmp"
            ;;
    esac

    # Move binary to correct location
    mv "/tmp/${binary_path}" "${SCRIPT_DIR}/${platform}/"

    # Set executable permissions for non-Windows binaries
    if [[ ! "${platform}" =~ "windows" ]]; then
        chmod +x "${SCRIPT_DIR}/${platform}/rg"
    fi

    # Cleanup
    rm -f "/tmp/${archive_name}"
    rm -rf "/tmp/ripgrep-${RIPGREP_VERSION}-"*

    echo "✓ ${platform} complete"
}

# macOS Intel
download_ripgrep \
    "amd64-darwin" \
    "https://github.com/BurntSushi/ripgrep/releases/download/${RIPGREP_VERSION}/ripgrep-${RIPGREP_VERSION}-x86_64-apple-darwin.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-x86_64-apple-darwin.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-x86_64-apple-darwin/rg"

# macOS Apple Silicon
download_ripgrep \
    "arm64-darwin" \
    "https://github.com/BurntSushi/ripgrep/releases/download/${RIPGREP_VERSION}/ripgrep-${RIPGREP_VERSION}-aarch64-apple-darwin.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-aarch64-apple-darwin.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-aarch64-apple-darwin/rg"

# Linux x64
download_ripgrep \
    "amd64-linux" \
    "https://github.com/BurntSushi/ripgrep/releases/download/${RIPGREP_VERSION}/ripgrep-${RIPGREP_VERSION}-x86_64-unknown-linux-musl.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-x86_64-unknown-linux-musl.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-x86_64-unknown-linux-musl/rg"

# Linux ARM64
download_ripgrep \
    "arm64-linux" \
    "https://github.com/BurntSushi/ripgrep/releases/download/${RIPGREP_VERSION}/ripgrep-${RIPGREP_VERSION}-aarch64-unknown-linux-gnu.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-aarch64-unknown-linux-gnu.tar.gz" \
    "ripgrep-${RIPGREP_VERSION}-aarch64-unknown-linux-gnu/rg"

# Windows x64
download_ripgrep \
    "amd64-windows" \
    "https://github.com/BurntSushi/ripgrep/releases/download/${RIPGREP_VERSION}/ripgrep-${RIPGREP_VERSION}-x86_64-pc-windows-msvc.zip" \
    "ripgrep-${RIPGREP_VERSION}-x86_64-pc-windows-msvc.zip" \
    "ripgrep-${RIPGREP_VERSION}-x86_64-pc-windows-msvc/rg.exe"

# Note: Windows ARM64 uses the same x64 binary via emulation
# The build system will use the amd64-windows binary for both architectures

echo "All ripgrep binaries downloaded successfully!"