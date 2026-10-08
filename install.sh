#!/bin/sh
set -e

# zentao-cli one-line installer for Linux, macOS & Windows (Git-Bash)
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/windosx/zentao-cli/main/install.sh | bash

REPO="windosx/zentao-cli"

# 1. Detect OS and Architecture
OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
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

case "$OS_RAW" in
  linux*)
    OS="linux"
    BIN_NAME="zentao"
    EXT="tar.gz"
    ;;
  darwin*)
    OS="darwin"
    BIN_NAME="zentao"
    EXT="tar.gz"
    ;;
  mingw*|msys*|cygwin*|windows*)
    OS="windows"
    BIN_NAME="zentao.exe"
    EXT="zip"
    ;;
  *)
    echo "Error: Unsupported operating system: $OS_RAW"
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
ARCHIVE_NAME="zentao-cli-${VERSION_CLEAN}-${OS}-${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ARCHIVE_NAME}"

# 3. Download and extract
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'zentao')
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

echo "==> Extracting archive..."
if [ "$EXT" = "zip" ]; then
  if command -v unzip >/dev/null 2>&1; then
    unzip -q "$TMP_DIR/$ARCHIVE_NAME" -d "$TMP_DIR"
  elif command -v tar >/dev/null 2>&1; then
    tar -xf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"
  elif command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -Command "Expand-Archive -Path '$TMP_DIR/$ARCHIVE_NAME' -DestinationPath '$TMP_DIR' -Force"
  else
    echo "Error: Neither unzip, tar, nor powershell found to extract zip archive."
    exit 1
  fi
else
  tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"
fi

# 4. Find binary
TARGET_BIN=$(find "$TMP_DIR" -type f -name "$BIN_NAME" | head -n 1)
if [ -z "$TARGET_BIN" ]; then
  echo "Error: Binary ${BIN_NAME} not found in archive."
  exit 1
fi

chmod +x "$TARGET_BIN"

# 5. Determine install directory
if [ -n "$INSTALL_DIR" ]; then
  mkdir -p "$INSTALL_DIR" 2>/dev/null || true
elif [ "$OS" = "windows" ]; then
  if [ -w "/usr/bin" ]; then
    INSTALL_DIR="/usr/bin"
  elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  elif [ -d "$HOME/bin" ] && [ -w "$HOME/bin" ]; then
    INSTALL_DIR="$HOME/bin"
  elif [ -d "$HOME/.local/bin" ] && [ -w "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
  else
    INSTALL_DIR="$HOME/bin"
    mkdir -p "$INSTALL_DIR"
  fi
else
  INSTALL_DIR="/usr/local/bin"
fi

# 6. Install binary
echo "==> Installing ${BIN_NAME} into ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TARGET_BIN" "$INSTALL_DIR/$BIN_NAME"
else
  if [ "$OS" = "windows" ]; then
    mkdir -p "$HOME/bin"
    INSTALL_DIR="$HOME/bin"
    mv "$TARGET_BIN" "$INSTALL_DIR/$BIN_NAME"
    echo "Notice: Installed to $INSTALL_DIR. Please ensure $INSTALL_DIR is in your PATH."
  else
    echo "==> Requesting sudo permissions to install into ${INSTALL_DIR}..."
    sudo mv "$TARGET_BIN" "$INSTALL_DIR/$BIN_NAME"
  fi
fi

echo "==> Successfully installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}!"
"$INSTALL_DIR/$BIN_NAME" version || true
