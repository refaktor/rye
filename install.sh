#!/bin/sh
set -e

REPO="refaktor/rye"
INSTALL_DIR="/usr/local/bin"

# 1. Detect Architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64)
        ARCH="x86_64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture $ARCH"
        exit 1
        ;;
esac

# 2. Detect OS
OS=$(uname -s)
case "$OS" in
    Linux)
        OS="Linux"
        ;;
    Darwin)
        OS="Darwin"
        ;;
    *)
        echo "Error: Unsupported OS $OS"
        exit 1
        ;;
esac

# 3. Determine Download Tool (wget or curl)
if command -v curl >/dev/null 2>&1; then
    FETCH="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
    FETCH="wget -qO-"
else
    echo "Error: Neither curl nor wget is installed."
    exit 1
fi

# 4. Fetch Latest Release Version Tag
echo "Fetching latest Rye release tag..."
TAG=$($FETCH "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG" ]; then
    echo "Error: Could not determine latest release version."
    exit 1
fi

TARBALL="rye_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$TARBALL"

# 5. Download and Extract
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading Rye $TAG for $OS/$ARCH..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL"
else
    wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$TARBALL"
fi

tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

# 6. Install Binary
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/rye" "$INSTALL_DIR/rye"
else
    echo "Installing to $INSTALL_DIR requires elevated privileges."
    sudo mv "$TMP_DIR/rye" "$INSTALL_DIR/rye"
fi

chmod +x "$INSTALL_DIR/rye"

echo ""
echo "Successfully installed Rye ($TAG) to $INSTALL_DIR/rye"
echo "Run 'rye' to start the REPL!"
