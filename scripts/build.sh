#!/usr/bin/env bash
# build.sh - Build the StackEye CLI with version injection
#
# This script calculates version information from git and injects it
# into the binary via ldflags. It's useful for standalone builds
# outside of the Makefile or GoReleaser.
#
# Usage:
#   ./scripts/build.sh                    # Build for current platform
#   ./scripts/build.sh -o /usr/local/bin  # Build with custom output directory
#   ./scripts/build.sh -v v1.2.3          # Override version
#   ./scripts/build.sh --help             # Show help
#
# Environment variables:
#   VERSION   - Override version (default: git describe)
#   COMMIT    - Override commit SHA (default: git rev-parse)
#   DATE      - Override build date (default: current UTC time)
#   BUILT_BY  - Override built by (default: "build.sh")

set -euo pipefail

# Configuration
BINARY_NAME="stackeye"
CMD_DIR="./cmd/stackeye"
OUTPUT_DIR="./bin"
VERSION_PKG="github.com/StackEye-IO/stackeye-cli/internal/version"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
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

# Show usage
usage() {
    cat << EOF
Build StackEye CLI with version injection

Usage: $0 [OPTIONS]

Options:
    -o, --output DIR    Output directory (default: ./bin)
    -v, --version VER   Override version string
    -h, --help          Show this help message

Environment Variables:
    VERSION     Override version (default: from git)
    COMMIT      Override commit SHA (default: from git)
    DATE        Override build date (default: current UTC)
    BUILT_BY    Override built by field (default: build.sh)

Examples:
    $0                          Build with auto-detected version
    $0 -v v1.0.0                Build with specific version
    $0 -o /usr/local/bin        Build to custom directory
    VERSION=dev $0              Build with VERSION=dev

EOF
    exit 0
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
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
}

# Get version from git, with dirty state detection
get_version() {
    if [[ -n "${VERSION:-}" ]]; then
        echo "$VERSION"
        return
    fi

    # Try to get version from git tags
    local version
    if version=$(git describe --tags --always 2>/dev/null); then
        # Check for dirty working directory
        if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
            version="${version}-dirty"
        fi
        echo "$version"
    else
        echo "dev"
    fi
}

# Get short commit SHA
get_commit() {
    if [[ -n "${COMMIT:-}" ]]; then
        echo "$COMMIT"
        return
    fi

    git rev-parse --short=7 HEAD 2>/dev/null || echo "none"
}

# Get build date in ISO 8601 format
get_date() {
    if [[ -n "${DATE:-}" ]]; then
        echo "$DATE"
        return
    fi

    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

# Get built by identifier
get_built_by() {
    echo "${BUILT_BY:-build.sh}"
}

# Main build function
build() {
    local version commit date built_by

    version=$(get_version)
    commit=$(get_commit)
    date=$(get_date)
    built_by=$(get_built_by)

    info "Building $BINARY_NAME..."
    info "  Version:   $version"
    info "  Commit:    $commit"
    info "  Date:      $date"
    info "  Built by:  $built_by"

    # Create output directory
    mkdir -p "$OUTPUT_DIR"

    # Build ldflags
    local ldflags="-s -w"
    ldflags="$ldflags -X '${VERSION_PKG}.Version=${version}'"
    ldflags="$ldflags -X '${VERSION_PKG}.Commit=${commit}'"
    ldflags="$ldflags -X '${VERSION_PKG}.Date=${date}'"
    ldflags="$ldflags -X '${VERSION_PKG}.BuiltBy=${built_by}'"

    # Build the binary
    local output_path="${OUTPUT_DIR}/${BINARY_NAME}"

    if go build -trimpath -ldflags "$ldflags" -o "$output_path" "$CMD_DIR"; then
        success "Built: $output_path"

        # Show binary info
        if command -v file &>/dev/null; then
            info "Binary type: $(file -b "$output_path")"
        fi

        # Show version output
        info "Version check:"
        "$output_path" version --short
    else
        error "Build failed"
        exit 1
    fi
}

# Entry point
main() {
    parse_args "$@"

    # Ensure we're in the project root
    if [[ ! -f "go.mod" ]]; then
        error "Must run from project root (where go.mod is located)"
        exit 1
    fi

    build
}

main "$@"
