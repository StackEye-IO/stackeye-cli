#!/usr/bin/env bash
# install.sh - Install StackEye CLI from GitHub Releases
#
# This script downloads and installs the StackEye CLI binary for your platform.
# It auto-detects OS and architecture, verifies checksums, and installs the binary.
#
# Usage:
#   curl -fsSL https://get.stackeye.io/cli | bash
#   curl -fsSL https://get.stackeye.io/cli | bash -s -- --version v1.0.0
#   curl -fsSL https://get.stackeye.io/cli | STACKEYE_INSTALL_DIR=/custom/path bash
#
# Environment variables:
#   STACKEYE_VERSION      - Install specific version (default: latest)
#   STACKEYE_INSTALL_DIR  - Installation directory (default: /usr/local/bin or ~/.local/bin)
#   STACKEYE_NO_VERIFY    - Skip checksum verification (not recommended)
#
# Supported platforms:
#   - Linux (amd64, arm64)
#   - macOS (amd64, arm64)

set -euo pipefail

# Configuration
GITHUB_REPO="StackEye-IO/stackeye-cli"
BINARY_NAME="stackeye"

# Colors for output (disabled if not a terminal)
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Print functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Check if a command exists
has_command() {
    command -v "$1" &>/dev/null
}

# Detect operating system
detect_os() {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux)
            echo "linux"
            ;;
        Darwin)
            echo "darwin"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            error "Windows is not supported by this installer."
            error "Please download the binary manually from:"
            error "  https://github.com/${GITHUB_REPO}/releases/latest"
            exit 1
            ;;
        *)
            error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
}

# Detect CPU architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            error "Supported architectures: x86_64 (amd64), aarch64 (arm64)"
            exit 1
            ;;
    esac
}

# Get latest version from GitHub API
get_latest_version() {
    local latest_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    local version

    if has_command curl; then
        version=$(curl -fsSL "$latest_url" 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    elif has_command wget; then
        version=$(wget -qO- "$latest_url" 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [[ -z "$version" ]]; then
        error "Failed to determine latest version from GitHub."
        error "Please check your internet connection or specify a version with --version."
        exit 1
    fi

    echo "$version"
}

# Download file
download_file() {
    local url="$1"
    local output="$2"

    info "Downloading: $url"

    if has_command curl; then
        curl -fsSL "$url" -o "$output"
    elif has_command wget; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local expected="$2"
    local actual

    if has_command sha256sum; then
        actual=$(sha256sum "$file" | awk '{print $1}')
    elif has_command shasum; then
        actual=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        warn "Neither sha256sum nor shasum found. Skipping checksum verification."
        return 0
    fi

    if [[ "$actual" != "$expected" ]]; then
        error "Checksum verification failed!"
        error "  Expected: $expected"
        error "  Actual:   $actual"
        return 1
    fi

    info "Checksum verified successfully"
    return 0
}

# Get checksum for a specific file from checksums.txt
get_expected_checksum() {
    local checksums_file="$1"
    local archive_name="$2"

    grep "${archive_name}" "$checksums_file" | awk '{print $1}'
}

# Determine installation directory
get_install_dir() {
    # Check for override
    if [[ -n "${STACKEYE_INSTALL_DIR:-}" ]]; then
        echo "$STACKEYE_INSTALL_DIR"
        return
    fi

    # Try /usr/local/bin first (requires sudo)
    if [[ -d "/usr/local/bin" && -w "/usr/local/bin" ]]; then
        echo "/usr/local/bin"
        return
    fi

    # Fall back to ~/.local/bin
    local local_bin="${HOME}/.local/bin"
    echo "$local_bin"
}

# Check if directory is in PATH
is_in_path() {
    local dir="$1"
    case ":${PATH}:" in
        *:${dir}:*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Install binary
install_binary() {
    local src="$1"
    local install_dir="$2"
    local dest="${install_dir}/${BINARY_NAME}"
    local need_sudo=false

    # Create directory if it doesn't exist
    if [[ ! -d "$install_dir" ]]; then
        if [[ "$install_dir" == "${HOME}/.local/bin" ]]; then
            mkdir -p "$install_dir"
        else
            if has_command sudo; then
                sudo mkdir -p "$install_dir"
                need_sudo=true
            else
                error "Cannot create directory $install_dir without sudo"
                exit 1
            fi
        fi
    fi

    # Check if we need sudo for the install directory
    if [[ ! -w "$install_dir" ]]; then
        if has_command sudo; then
            need_sudo=true
        else
            error "Cannot write to $install_dir and sudo is not available"
            error "Please set STACKEYE_INSTALL_DIR to a writable directory"
            exit 1
        fi
    fi

    # Install the binary
    if $need_sudo; then
        info "Installing to ${dest} (requires sudo)"
        sudo install -m 755 "$src" "$dest"
    else
        info "Installing to ${dest}"
        install -m 755 "$src" "$dest"
    fi

    success "StackEye CLI installed to ${dest}"

    # Check if install dir is in PATH
    if ! is_in_path "$install_dir"; then
        warn "Installation directory is not in your PATH"
        warn "Add this to your shell profile:"
        echo ""
        echo "  export PATH=\"\$PATH:${install_dir}\""
        echo ""
    fi
}

# Cleanup temporary files
cleanup() {
    if [[ -n "${_INSTALL_TMPDIR:-}" && -d "$_INSTALL_TMPDIR" ]]; then
        rm -rf "$_INSTALL_TMPDIR"
    fi
}

# Show usage
usage() {
    cat << EOF
Install StackEye CLI

Usage: $0 [OPTIONS]

Options:
    -v, --version VERSION   Install specific version (default: latest)
    -d, --dir DIR           Installation directory (default: /usr/local/bin or ~/.local/bin)
    -h, --help              Show this help message

Environment Variables:
    STACKEYE_VERSION        Install specific version
    STACKEYE_INSTALL_DIR    Installation directory
    STACKEYE_NO_VERIFY      Skip checksum verification (not recommended)

Examples:
    # Install latest version
    curl -fsSL https://get.stackeye.io/cli | bash

    # Install specific version
    curl -fsSL https://get.stackeye.io/cli | bash -s -- --version v1.0.0

    # Install to custom directory
    curl -fsSL https://get.stackeye.io/cli | STACKEYE_INSTALL_DIR=~/bin bash

EOF
    exit 0
}

# Main function
main() {
    local version="${STACKEYE_VERSION:-}"
    local install_dir=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                if [[ -z "${2:-}" ]]; then
                    error "--version requires a version argument (e.g., --version v1.0.0)"
                    exit 1
                fi
                version="$2"
                shift 2
                ;;
            -d|--dir)
                if [[ -z "${2:-}" ]]; then
                    error "--dir requires a directory argument (e.g., --dir ~/bin)"
                    exit 1
                fi
                install_dir="$2"
                shift 2
                ;;
            -h|--help)
                usage
                ;;
            *)
                error "Unknown option: $1"
                usage
                ;;
        esac
    done

    # Override install dir if set via argument
    if [[ -n "$install_dir" ]]; then
        export STACKEYE_INSTALL_DIR="$install_dir"
    fi

    echo ""
    echo "  StackEye CLI Installer"
    echo "  ======================"
    echo ""

    # Detect platform
    local os arch
    os=$(detect_os)
    arch=$(detect_arch)
    info "Detected platform: ${os}/${arch}"

    # Get version
    if [[ -z "$version" ]]; then
        info "Determining latest version..."
        version=$(get_latest_version)
    fi
    info "Version: ${version}"

    # Strip leading 'v' from version for archive name if present
    local version_no_v="${version#v}"

    # Construct archive name (matches GoReleaser output)
    local archive_name="${BINARY_NAME}_${version_no_v}_${os}_${arch}.tar.gz"
    local checksums_name="checksums.txt"

    # Construct download URLs
    local base_url="https://github.com/${GITHUB_REPO}/releases/download/${version}"
    local archive_url="${base_url}/${archive_name}"
    local checksums_url="${base_url}/${checksums_name}"

    # Create temporary directory
    _INSTALL_TMPDIR=$(mktemp -d)
    trap cleanup EXIT

    local archive_path="${_INSTALL_TMPDIR}/${archive_name}"
    local checksums_path="${_INSTALL_TMPDIR}/${checksums_name}"

    # Download checksums
    download_file "$checksums_url" "$checksums_path"

    # Download archive
    download_file "$archive_url" "$archive_path"

    # Verify checksum (unless disabled)
    if [[ -z "${STACKEYE_NO_VERIFY:-}" ]]; then
        local expected_checksum
        expected_checksum=$(get_expected_checksum "$checksums_path" "$archive_name")
        if [[ -z "$expected_checksum" ]]; then
            error "Could not find checksum for ${archive_name}"
            exit 1
        fi
        verify_checksum "$archive_path" "$expected_checksum" || exit 1
    else
        warn "Checksum verification skipped (STACKEYE_NO_VERIFY is set)"
    fi

    # Extract archive
    info "Extracting archive..."
    tar -xzf "$archive_path" -C "$_INSTALL_TMPDIR"

    # Find the binary in the extracted content
    local binary_path="${_INSTALL_TMPDIR}/${BINARY_NAME}"
    if [[ ! -f "$binary_path" ]]; then
        error "Binary not found in archive"
        exit 1
    fi

    # Determine installation directory
    install_dir=$(get_install_dir)

    # Install
    install_binary "$binary_path" "$install_dir"

    # Verify installation
    echo ""
    if has_command "$BINARY_NAME"; then
        info "Verifying installation..."
        "$BINARY_NAME" version
    else
        info "Installation complete. You may need to restart your shell or run:"
        echo ""
        echo "  export PATH=\"\$PATH:${install_dir}\""
        echo ""
    fi

    echo ""
    success "StackEye CLI has been installed successfully!"
    echo ""
    echo "  Get started:"
    echo "    stackeye login    # Authenticate with StackEye"
    echo "    stackeye --help   # See all available commands"
    echo ""
}

main "$@"
