#!/bin/sh

set -e

REPO="piyush-gambhir/cubeapm-cli"
BINARY="cubeapm"
PROJECT="cubeapm-cli"

if [ -t 1 ]; then
    GREEN='\033[0;32m'
    BLUE='\033[0;34m'
    RED='\033[0;31m'
    YELLOW='\033[0;33m'
    NC='\033[0m'
else
    GREEN='' BLUE='' RED='' YELLOW='' NC=''
fi

info() { printf "${BLUE}==>${NC} %s\n" "$1"; }
success() { printf "${GREEN}==>${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}==>${NC} %s\n" "$1"; }
error() { printf "${RED}error:${NC} %s\n" "$1" >&2; exit 1; }

command -v curl >/dev/null 2>&1 || error "curl is required but not installed"
command -v tar >/dev/null 2>&1 || error "tar is required but not installed"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    linux) ;;
    darwin) ;;
    *) error "Unsupported OS: $OS" ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

if [ -z "$VERSION" ]; then
    info "Fetching latest version..."
    VERSION=$(curl -sSfI "https://github.com/${REPO}/releases/latest" | tr -d '\r' | sed -n 's#^[Ll]ocation: .*/releases/tag/v\([0-9][0-9A-Za-z.-]*\)$#\1#p') || error "Failed to fetch latest version. Set VERSION env var manually."
    [ -z "$VERSION" ] && error "Could not determine latest version. Set VERSION env var manually."
fi

info "Installing ${BINARY} v${VERSION} (${OS}/${ARCH})"

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

ARCHIVE="${PROJECT}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"
info "Downloading ${URL}..."
curl -sSfL "$URL" -o "${TMP_DIR}/${ARCHIVE}" || error "Download failed. Check that version v${VERSION} exists at https://github.com/${REPO}/releases"

info "Extracting..."
tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"

mkdir -p "$INSTALL_DIR" 2>/dev/null || true
if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
    info "Elevated permissions required to install to ${INSTALL_DIR}"
    sudo mkdir -p "$INSTALL_DIR"
    sudo mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi
chmod +x "${INSTALL_DIR}/${BINARY}"

if command -v "$BINARY" >/dev/null 2>&1; then
    success "Successfully installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
    "${INSTALL_DIR}/${BINARY}" version
else
    success "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
    warn "${INSTALL_DIR} may not be in your PATH"
fi
