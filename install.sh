#!/bin/bash
# Copyright (c) 2025 JoeGlenn1213
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

# LGH Installer Script for macOS/Linux
# Usage: curl -sSL https://raw.githubusercontent.com/JoeGlenn1213/lgh/main/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="lgh"
REPO="JoeGlenn1213/lgh"

# Resolve the version to install: LGH_VERSION env override wins, otherwise
# ask the GitHub API for the latest release tag (stripped of the leading v).
if [ -n "$LGH_VERSION" ]; then
    VERSION="$LGH_VERSION"
else
    VERSION=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        echo -e "${RED}✗ Could not resolve the latest release (GitHub API unreachable?).${NC}"
        echo -e "${RED}  Set LGH_VERSION=<x.y.z> and retry, or install from source:${NC}"
        echo -e "${RED}  git clone https://github.com/$REPO && cd lgh && make build${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║            LGH (LocalGitHub) Installer v${VERSION}                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}✗ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

if [ "$OS" != "darwin" ] && [ "$OS" != "linux" ]; then
    echo -e "${RED}✗ Unsupported OS: $OS${NC}"
    exit 1
fi

echo -e "${YELLOW}ℹ Detected: ${OS}/${ARCH}${NC}"

# Release assets are named lgh-v<tag>-<os>-<arch> (see .github/workflows/release.yml).
ASSET="lgh-v${VERSION}-${OS}-${ARCH}"

# Check if already installed
if command -v lgh &> /dev/null; then
    EXISTING_VERSION=$(lgh --version 2>&1 | head -1)
    echo -e "${YELLOW}⚠ LGH is already installed: $EXISTING_VERSION${NC}"
    read -p "Do you want to reinstall? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}ℹ Installation cancelled.${NC}"
        exit 0
    fi
fi

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Check for local binary matching current architecture (dev convenience:
# `make release` produces lgh-<version>-<os>-<arch> without the v prefix)
LOCAL_BINARY=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -f "$SCRIPT_DIR/dist/lgh" ]; then
    LOCAL_BINARY="$SCRIPT_DIR/dist/lgh"
    echo -e "${GREEN}✓ Found local binary: $LOCAL_BINARY${NC}"
else
    LOCAL_BINARY=$(ls "$SCRIPT_DIR"/dist/lgh-*-"${OS}-${ARCH}" 2>/dev/null | head -1 || true)
    if [ -n "$LOCAL_BINARY" ]; then
        echo -e "${GREEN}✓ Found local binary: $LOCAL_BINARY${NC}"
    fi
fi

if [ -n "$LOCAL_BINARY" ]; then
    # Use local binary
    cp "$LOCAL_BINARY" "$TMP_DIR/lgh"
else
    # Download from GitHub
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/v${VERSION}/${ASSET}"
    CHECKSUM_URL="${DOWNLOAD_URL}.sha256"

    echo -e "${BLUE}ℹ Downloading LGH v${VERSION} (${ASSET}) from GitHub...${NC}"

    if command -v curl &> /dev/null; then
        curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/lgh" || {
            echo -e "${RED}✗ Download failed. Please check your internet connection.${NC}"
            exit 1
        }
        # Try to download and verify checksum
        if curl -sL "$CHECKSUM_URL" -o "$TMP_DIR/checksum.txt" 2>/dev/null; then
            EXPECTED_SHA=$(awk '{print $1}' "$TMP_DIR/checksum.txt")
            if [ -n "$EXPECTED_SHA" ]; then
                echo -e "${BLUE}ℹ Verifying checksum...${NC}"
                if command -v sha256sum &> /dev/null; then
                    ACTUAL_SHA=$(sha256sum "$TMP_DIR/lgh" | awk '{print $1}')
                elif command -v shasum &> /dev/null; then
                    ACTUAL_SHA=$(shasum -a 256 "$TMP_DIR/lgh" | awk '{print $1}')
                else
                    echo -e "${YELLOW}⚠ sha256sum/shasum not found, skipping checksum verification${NC}"
                    ACTUAL_SHA=""
                fi
                
                if [ -n "$ACTUAL_SHA" ]; then
                    if [ "$EXPECTED_SHA" = "$ACTUAL_SHA" ]; then
                        echo -e "${GREEN}✓ Checksum verified${NC}"
                    else
                        echo -e "${RED}✗ Checksum mismatch!${NC}"
                        echo -e "${RED}  Expected: $EXPECTED_SHA${NC}"
                        echo -e "${RED}  Actual:   $ACTUAL_SHA${NC}"
                        echo -e "${RED}  This could indicate a corrupted download or supply chain attack.${NC}"
                        exit 1
                    fi
                fi
            else
                echo -e "${YELLOW}⚠ No checksum found for ${ASSET}, skipping verification${NC}"
            fi
        else
            echo -e "${YELLOW}⚠ Could not download checksums, skipping verification${NC}"
            echo -e "${YELLOW}  For security, consider verifying manually after installation${NC}"
        fi
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/lgh" || {
            echo -e "${RED}✗ Download failed. Please check your internet connection.${NC}"
            exit 1
        }
        echo -e "${YELLOW}⚠ wget does not support checksum verification in this script${NC}"
    else
        echo -e "${RED}✗ Neither curl nor wget found. Please install one of them.${NC}"
        exit 1
    fi
fi

# Make executable
chmod +x "$TMP_DIR/lgh"

# Verify binary runs
if ! "$TMP_DIR/lgh" --version &> /dev/null; then
    echo -e "${RED}✗ Downloaded binary is not valid or not compatible with this system.${NC}"
    exit 1
fi

# Additional verification: check if it's actually an LGH binary
VERSION_OUTPUT=$("$TMP_DIR/lgh" --version 2>&1)
if ! echo "$VERSION_OUTPUT" | grep -q "LGH"; then
    echo -e "${RED}✗ Binary does not appear to be LGH.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Binary verified: $VERSION_OUTPUT${NC}"

# Install
echo -e "${BLUE}ℹ Installing to $INSTALL_DIR...${NC}"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/lgh" "$INSTALL_DIR/$BINARY_NAME"
else
    echo -e "${YELLOW}⚠ Need sudo permission to install to $INSTALL_DIR${NC}"
    sudo mv "$TMP_DIR/lgh" "$INSTALL_DIR/$BINARY_NAME"
fi

# Verify installation
if command -v lgh &> /dev/null; then
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║               ✓ LGH installed successfully!                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    lgh --version
    echo
    echo -e "${BLUE}Quick Start:${NC}"
    echo "  lgh init          # Initialize LGH environment"
    echo "  lgh serve         # Start the HTTP server"
    echo "  lgh add .         # Add current directory"
    echo "  git push lgh main # Push to local GitHub!"
    echo
    echo -e "${BLUE}For network sharing (with auth):${NC}"
    echo "  lgh auth setup    # Set up username/password"
    echo "  lgh serve --bind 0.0.0.0 --read-only"
else
    echo -e "${RED}✗ Installation failed.${NC}"
    exit 1
fi

