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

# LGH Uninstaller Script
# Usage: curl -sSL https://raw.githubusercontent.com/JoeGlenn1213/lgh/main/uninstall.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="lgh"
DATA_DIR="$HOME/.localgithub"

echo -e "${BLUE}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║            LGH (LocalGitHub) Uninstaller                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if installed
if ! command -v lgh &> /dev/null; then
    echo -e "${YELLOW}⚠ LGH is not installed.${NC}"
    exit 0
fi

echo -e "${YELLOW}ℹ Found LGH installation:${NC}"
lgh --version
echo

# Check for running server
if lgh status 2>/dev/null | grep -q "RUNNING"; then
    echo -e "${YELLOW}⚠ LGH server is running. Stopping...${NC}"
    pkill -f "lgh serve" 2>/dev/null || true
    sleep 1
fi

# Ask about data directory
if [ -d "$DATA_DIR" ]; then
    echo -e "${YELLOW}⚠ Found LGH data directory: $DATA_DIR${NC}"
    read -p "Do you want to remove data directory (contains your repos)? [y/N] " -n 1 -r
    echo
    REMOVE_DATA=$REPLY
fi

# Confirm uninstall
read -p "Are you sure you want to uninstall LGH? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}ℹ Uninstallation cancelled.${NC}"
    exit 0
fi

# Remove binary
echo -e "${BLUE}ℹ Removing LGH binary...${NC}"
if [ -w "$INSTALL_DIR/$BINARY_NAME" ]; then
    rm -f "$INSTALL_DIR/$BINARY_NAME"
else
    sudo rm -f "$INSTALL_DIR/$BINARY_NAME"
fi
echo -e "${GREEN}✓ Removed $INSTALL_DIR/$BINARY_NAME${NC}"

# Remove data directory if requested
if [[ $REMOVE_DATA =~ ^[Yy]$ ]] && [ -d "$DATA_DIR" ]; then
    echo -e "${BLUE}ℹ Removing data directory...${NC}"
    rm -rf "$DATA_DIR"
    echo -e "${GREEN}✓ Removed $DATA_DIR${NC}"
fi

echo -e "${GREEN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║            ✓ LGH uninstalled successfully!                   ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
