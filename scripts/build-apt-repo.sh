#!/usr/bin/env bash
# build-apt-repo.sh - Build APT repository structure from .deb files
#
# Usage: ./scripts/build-apt-repo.sh <version-tag> <gpg-fingerprint> <dist-dir>
#
# Produces:
#   apt-repo/
#   ├── dists/stable/
#   │   ├── main/binary-amd64/Packages, Packages.gz
#   │   ├── main/binary-arm64/Packages, Packages.gz
#   │   ├── Release
#   │   ├── Release.gpg
#   │   └── InRelease
#   └── pool/main/s/stackeye/
#       ├── stackeye_X.Y.Z_linux_amd64.deb
#       └── stackeye_X.Y.Z_linux_arm64.deb
#
# Requires: dpkg-scanpackages, apt-ftparchive, gpg2, gzip

set -euo pipefail

VERSION_TAG="${1:?Usage: $0 <version-tag> <gpg-fingerprint> <dist-dir>}"
GPG_FINGERPRINT="${2:?Usage: $0 <version-tag> <gpg-fingerprint> <dist-dir>}"
DIST_DIR="${3:?Usage: $0 <version-tag> <gpg-fingerprint> <dist-dir>}"

# Strip leading 'v' from tag to get version number
VERSION="${VERSION_TAG#v}"

REPO_ROOT="apt-repo"
POOL_DIR="${REPO_ROOT}/pool/main/s/stackeye"
DISTS_DIR="${REPO_ROOT}/dists/stable"

echo "Building APT repository for version ${VERSION}..."

# Clean previous build
rm -rf "${REPO_ROOT}"

# Create directory structure
mkdir -p "${POOL_DIR}"
mkdir -p "${DISTS_DIR}/main/binary-amd64"
mkdir -p "${DISTS_DIR}/main/binary-arm64"

# Copy .deb files into pool
DEBS_FOUND=0
for arch in amd64 arm64; do
    DEB_FILE="${DIST_DIR}/stackeye_${VERSION}_linux_${arch}.deb"
    if [[ -f "${DEB_FILE}" ]]; then
        cp "${DEB_FILE}" "${POOL_DIR}/"
        DEBS_FOUND=$((DEBS_FOUND + 1))
        echo "  Found: ${DEB_FILE}"
    else
        echo "  Warning: ${DEB_FILE} not found, skipping ${arch}"
    fi
done

if [[ "${DEBS_FOUND}" -eq 0 ]]; then
    echo "Error: No .deb files found in ${DIST_DIR}"
    exit 1
fi

# Generate Packages files for each architecture
# dpkg-scanpackages must run from the repo root so Filename: paths are correct
cd "${REPO_ROOT}"

for arch in amd64 arm64; do
    PACKAGES_DIR="dists/stable/main/binary-${arch}"

    # Check if any .deb files exist for this architecture
    if ls pool/main/s/stackeye/*_${arch}.deb 1>/dev/null 2>&1; then
        dpkg-scanpackages --arch "${arch}" pool/ > "${PACKAGES_DIR}/Packages"
        gzip -9c "${PACKAGES_DIR}/Packages" > "${PACKAGES_DIR}/Packages.gz"
        echo "  Generated Packages for ${arch}"
    else
        # Create empty Packages file for missing architectures
        touch "${PACKAGES_DIR}/Packages"
        gzip -9c "${PACKAGES_DIR}/Packages" > "${PACKAGES_DIR}/Packages.gz"
        echo "  Generated empty Packages for ${arch} (no .deb found)"
    fi
done

# Generate Release file
cd dists/stable

apt-ftparchive release \
    -o APT::FTPArchive::Release::Origin="StackEye" \
    -o APT::FTPArchive::Release::Label="StackEye CLI" \
    -o APT::FTPArchive::Release::Suite="stable" \
    -o APT::FTPArchive::Release::Codename="stable" \
    -o APT::FTPArchive::Release::Architectures="amd64 arm64" \
    -o APT::FTPArchive::Release::Components="main" \
    . > Release

echo "  Generated Release file"

# Sign the Release file
gpg2 --batch --local-user "${GPG_FINGERPRINT}" --armor --detach-sign --output Release.gpg Release
gpg2 --batch --local-user "${GPG_FINGERPRINT}" --clearsign --output InRelease Release

echo "  Signed Release (Release.gpg + InRelease)"

cd ../../..

echo ""
echo "APT repository built at ${REPO_ROOT}/"
echo "Structure:"
find "${REPO_ROOT}" -type f | sort | while read -r f; do
    echo "  ${f}"
done
