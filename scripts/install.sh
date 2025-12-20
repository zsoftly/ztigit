#!/usr/bin/env bash
#
# One-liner installer for ztigit on macOS, Linux, WSL.
#
# Usage:
#   curl -fsSL https://github.com/zsoftly/ztigit/releases/latest/download/install.sh | bash
#

set -euo pipefail

REPO="zsoftly/ztigit"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="ztigit"

# Colors
info() { printf "\033[0;36m  %s\033[0m\n" "$1"; }
ok() { printf "\033[0;32m  [OK] %s\033[0m\n" "$1"; }
error() { printf "\033[0;31m  [ERROR] %s\033[0m\n" "$1"; exit 1; }

echo ""
echo "  ztigit installer"
echo "  ----------------"
echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    mingw*|msys*|cygwin*) error "Use PowerShell installer on Windows" ;;
    *) error "Unsupported OS: $OS" ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

ASSET_NAME="ztigit-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ASSET_NAME}"

# Download
info "Downloading ${ASSET_NAME}..."
TMP_FILE=$(mktemp)
trap "rm -f '$TMP_FILE'" EXIT

if command -v curl &> /dev/null; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
elif command -v wget &> /dev/null; then
    wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
else
    error "curl or wget required"
fi

chmod +x "$TMP_FILE"

# Install
DEST_PATH="${INSTALL_DIR}/${BINARY_NAME}"
info "Installing to ${DEST_PATH}..."

if [[ -w "$INSTALL_DIR" ]]; then
    mv "$TMP_FILE" "$DEST_PATH"
else
    sudo mv "$TMP_FILE" "$DEST_PATH"
    sudo chmod +x "$DEST_PATH"
fi

# Verify
if command -v ztigit &> /dev/null; then
    VERSION=$(ztigit --version 2>&1 || echo "ztigit (version unavailable)")
    echo ""
    ok "Installed successfully!"
    echo "  $VERSION"
else
    echo ""
    ok "Installed to $DEST_PATH"
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo "  Add to PATH:"
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
fi

echo ""
echo "  Usage:"
echo "    ztigit mirror <org> --provider github"
echo "    ztigit --help"
echo ""
