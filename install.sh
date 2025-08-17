#!/bin/bash

set -euo pipefail

# Script to install secret CLI from GitHub releases

REPO="Avdushin/secret"
VERSION="0.0.1"  # Change to 'latest' if using GitHub API for latest tag
BINARY_NAME="secret"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
if [ "$OS" = "darwin" ]; then
    OS="darwin"
elif [ "$OS" = "linux" ]; then
    OS="linux"
else
    echo "Unsupported OS: $OS. For Windows, download manually from releases."
    exit 1
fi

# Detect architecture
ARCH="$(uname -m)"
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Construct download URL
if [ "$OS" = "windows" ]; then
    EXT=".exe"
else
    EXT=""
fi
FILE_NAME="${BINARY_NAME}-${VERSION}-${OS}-${ARCH}${EXT}"
RELEASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILE_NAME}"

# Download the binary
echo "Downloading ${FILE_NAME} from ${RELEASE_URL}..."
curl -L -o "${BINARY_NAME}" "${RELEASE_URL}"

# Make executable (skip for Windows .exe)
if [ "$OS" != "windows" ]; then
    chmod +x "${BINARY_NAME}"
fi

# Move to install dir
echo "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
sudo mv "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

echo "Installation complete! Run 'secret --help' to get started."