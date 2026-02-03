#!/usr/bin/env bash
# build-rpm-repo.sh - Build YUM/DNF repository structure from .rpm files
#
# Usage: ./scripts/build-rpm-repo.sh <version-tag> <gpg-fingerprint> <dist-dir>
#
# Produces:
#   yum-repo/stable/
#   ├── x86_64/
#   │   ├── stackeye_X.Y.Z_linux_amd64.rpm
#   │   └── repodata/
#   │       ├── repomd.xml
#   │       ├── repomd.xml.asc
#   │       └── *.gz
#   └── aarch64/
#       ├── stackeye_X.Y.Z_linux_arm64.rpm
#       └── repodata/
#
# Requires: createrepo_c, gpg2

set -euo pipefail

VERSION_TAG="${1:?Usage: $0 <version-tag> <gpg-fingerprint> <dist-dir>}"
GPG_FINGERPRINT="${2:?Usage: $0 <version-tag> <gpg-fingerprint> <dist-dir>}"
DIST_DIR="${3:?Usage: $0 <version-tag> <gpg-fingerprint> <dist-dir>}"

# Strip leading 'v' from tag to get version number
VERSION="${VERSION_TAG#v}"

REPO_ROOT="yum-repo/stable"

# Map GoReleaser arch names to RPM arch names
declare -A ARCH_MAP=(
    ["amd64"]="x86_64"
    ["arm64"]="aarch64"
)

echo "Building YUM repository for version ${VERSION}..."

# Clean previous build
rm -rf "yum-repo"

RPMS_FOUND=0
for go_arch in amd64 arm64; do
    rpm_arch="${ARCH_MAP[${go_arch}]}"
    ARCH_DIR="${REPO_ROOT}/${rpm_arch}"
    mkdir -p "${ARCH_DIR}"

    RPM_FILE="${DIST_DIR}/stackeye_${VERSION}_linux_${go_arch}.rpm"
    if [[ -f "${RPM_FILE}" ]]; then
        cp "${RPM_FILE}" "${ARCH_DIR}/"
        RPMS_FOUND=$((RPMS_FOUND + 1))
        echo "  Found: ${RPM_FILE} -> ${ARCH_DIR}/"
    else
        echo "  Warning: ${RPM_FILE} not found, skipping ${rpm_arch}"
    fi

    # Create repo metadata even for empty directories (valid empty repo)
    createrepo_c "${ARCH_DIR}"
    echo "  Generated repodata for ${rpm_arch}"

    # Sign repomd.xml
    gpg2 --batch --local-user "${GPG_FINGERPRINT}" --armor --detach-sign \
        --output "${ARCH_DIR}/repodata/repomd.xml.asc" \
        "${ARCH_DIR}/repodata/repomd.xml"
    echo "  Signed repomd.xml for ${rpm_arch}"
done

if [[ "${RPMS_FOUND}" -eq 0 ]]; then
    echo "Error: No .rpm files found in ${DIST_DIR}"
    exit 1
fi

echo ""
echo "YUM repository built at yum-repo/"
echo "Structure:"
find "yum-repo" -type f | sort | while read -r f; do
    echo "  ${f}"
done
