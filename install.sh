#!/bin/sh
set -e

# zentao-cli one-line installer for Linux & macOS
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/windosx/zentao-cli/main/install.sh | bash

REPO="windosx/zentao-cli"
BIN_NAME="zentao"
INSTALL_DIR="/usr/local/bin"

# 1. Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin)
    ;;
  *)
    echo "Error: Unsupported operating system: $OS"
    exit 1
    ;;
esac

echo "==> Installing ${BIN_NAME} for ${OS}/${ARCH}..."

# 2. Get latest release tag
LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_TAG" ]; then
  LATEST_TAG=$(curl -s -L -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" | sed 's#.*/tag/##')
fi

if [ -z "$LATEST_TAG" ]; then
  echo "Error: Could not determine latest release version."
  exit 1
fi

VERSION_CLEAN=$(echo "$LATEST_TAG" | sed 's/^v//')
ARCHIVE_NAME="zentao-cli-${VERSION_CLEAN}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ARCHIVE_NAME}"

# 3. Download and extract
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "==> Downloading ${DOWNLOAD_URL}..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"
else
  echo "Error: Neither curl nor wget found."
  exit 1
fi

tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"

# 4. Install binary
TARGET_BIN=$(find "$TMP_DIR" -type f -name "$BIN_NAME" | head -n 1)
if [ -z "$TARGET_BIN" ]; then
  echo "Error: Binary ${BIN_NAME} not found in archive."
  exit 1
fi

chmod +x "$TARGET_BIN"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TARGET_BIN" "$INSTALL_DIR/$BIN_NAME"
else
  echo "==> Requesting sudo permissions to install into ${INSTALL_DIR}..."
  sudo mv "$TARGET_BIN" "$INSTALL_DIR/$BIN_NAME"
fi

echo "==> Successfully installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}!"
"$INSTALL_DIR/$BIN_NAME" version || true
