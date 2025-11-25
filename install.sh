#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# GitHub repository
REPO="njenia/envgrd"
BINARY_NAME="envgrd"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    case "$OS" in
        linux)
            PLATFORM="linux"
            EXT="tar.gz"
            # Only amd64 Linux builds are available in releases
            # ARM64 Linux and other platforms should build from source
            if [ "$ARCH" != "amd64" ]; then
                echo -e "${YELLOW}Note: ${ARCH} Linux binaries are not available in releases.${NC}"
                echo "Please build from source:"
                echo "  git clone https://github.com/$REPO.git"
                echo "  cd envgrd && make build"
                exit 0
            fi
            ;;
        darwin)
            PLATFORM="darwin"
            EXT="tar.gz"
            ;;
        *)
            echo -e "${YELLOW}Note: Pre-built binaries for $OS are not available in releases.${NC}"
            echo "Please build from source:"
            echo "  git clone https://github.com/$REPO.git"
            echo "  cd envgrd && make build"
            exit 0
            ;;
    esac
}

# Download and install
install() {
    detect_platform
    
    VERSION="${VERSION:-latest}"
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/$REPO/releases/latest/download/${BINARY_NAME}-${PLATFORM}-${ARCH}.${EXT}"
    else
        DOWNLOAD_URL="https://github.com/$REPO/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}-${ARCH}.${EXT}"
    fi
    
    echo -e "${GREEN}Installing ${BINARY_NAME}...${NC}"
    echo "Platform: ${PLATFORM}-${ARCH}"
    echo "Download URL: $DOWNLOAD_URL"
    
    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT
    
    # Download
    echo "Downloading..."
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/${BINARY_NAME}.${EXT}"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/${BINARY_NAME}.${EXT}"
    else
        echo -e "${RED}Error: curl or wget is required${NC}"
        exit 1
    fi
    
    # Extract
    echo "Extracting..."
    cd "$TMP_DIR"
    if [ "$EXT" = "tar.gz" ]; then
        tar -xzf "${BINARY_NAME}.${EXT}"
    else
        unzip -q "${BINARY_NAME}.${EXT}"
    fi
    
    # Install
    echo "Installing to $INSTALL_DIR..."
    if [ ! -w "$INSTALL_DIR" ]; then
        echo "Requires sudo to install to $INSTALL_DIR"
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        mv "$BINARY_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        INSTALLED_VERSION=$($BINARY_NAME version 2>/dev/null || echo "unknown")
        echo -e "${GREEN}✓ Successfully installed ${BINARY_NAME}${NC}"
        echo "Version: $INSTALLED_VERSION"
        echo "Run '${BINARY_NAME} scan' to get started!"
    else
        echo -e "${YELLOW}Warning: ${BINARY_NAME} was installed but not found in PATH${NC}"
        echo "Make sure $INSTALL_DIR is in your PATH"
    fi
}

# Run installation
install

