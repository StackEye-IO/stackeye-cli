# Release Checklist

This document describes the end-to-end process for cutting a new StackEye CLI release. The release pipeline is fully automated via GitHub Actions — your job is to prepare the repository, push a tag, and verify the results.

## Distribution Channels

A single `git push --tags` triggers all of these:

| Channel | Destination | Mechanism |
|---------|-------------|-----------|
| GitHub Releases | github.com/StackEye-IO/stackeye-cli/releases | GoReleaser |
| CDN Archives | releases.stackeye.io/cli/vX.Y.Z/ | Wasabi S3 + CloudFlare |
| APT Repository | releases.stackeye.io/dists/stable/ | `scripts/build-apt-repo.sh` |
| YUM Repository | releases.stackeye.io/yum/stable/ | `scripts/build-rpm-repo.sh` |
| Homebrew Tap | StackEye-IO/homebrew-tap | GoReleaser brews |
| Scoop Bucket | StackEye-IO/scoop-bucket | GoReleaser scoops |
| Docker (GHCR) | ghcr.io/stackeye-io/stackeye-cli | GoReleaser dockers |
| Installer Scripts | releases.stackeye.io/install.sh, install.ps1 | S3 upload |
| GPG Signatures | releases.stackeye.io/gpg-key.asc | S3 upload |

---

## Pre-Release Checklist

Complete these items before tagging.

- [ ] **All CI checks pass on `main`** — verify at [Actions](https://github.com/StackEye-IO/stackeye-cli/actions/workflows/ci.yml)
- [ ] **Run full local validation**:
  ```bash
  make validate
  ```
- [ ] **CHANGELOG updated** — add a section for the new version summarizing features, fixes, and breaking changes (GoReleaser auto-generates one from commits, but a curated summary is preferred for major releases)
- [ ] **Version string is correct** — the version comes from the git tag, not from source. Verify `internal/version` uses ldflags injection (no hardcoded version)
- [ ] **Dependencies are current** — check for any pending Dependabot PRs that should be merged first
- [ ] **No unmerged breaking changes** — confirm all intended PRs are merged to `main`
- [ ] **Test snapshot build locally**:
  ```bash
  make release-dry-run
  ```
  Verify `dist/` contains:
  - `stackeye_*_darwin_amd64.tar.gz`
  - `stackeye_*_darwin_arm64.tar.gz`
  - `stackeye_*_linux_amd64.tar.gz`
  - `stackeye_*_linux_arm64.tar.gz`
  - `stackeye_*_windows_amd64.zip`
  - `stackeye_*_linux_amd64.deb`
  - `stackeye_*_linux_arm64.deb`
  - `stackeye_*_linux_amd64.rpm`
  - `stackeye_*_linux_arm64.rpm`
  - `checksums.txt`
  - `checksums.txt.sig` (requires GPG key)

---

## Release Process

### 1. Create and Push a Tag

Tags follow [Semantic Versioning](https://semver.org/) with a `v` prefix.

```bash
# Stable release
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0

# Pre-release (GoReleaser marks these automatically)
git tag -a v1.2.0-rc.1 -m "Release v1.2.0-rc.1"
git push origin v1.2.0-rc.1
```

### 2. Monitor the Release Workflow

The tag push triggers `.github/workflows/release.yml`. Monitor progress at [Actions](https://github.com/StackEye-IO/stackeye-cli/actions/workflows/release.yml).

The workflow performs these steps in order:

1. Checkout repository (full history for changelog)
2. Checkout `stackeye-go-sdk` (private dependency)
3. Setup Go and dependencies
4. Import GPG signing key
5. Setup Docker Buildx + QEMU (multi-arch)
6. Login to GHCR
7. **GoReleaser** — builds binaries, archives, .deb/.rpm, Docker images, GitHub Release, Homebrew formula, Scoop manifest, GPG-signed checksums
8. Upload artifacts to Wasabi S3
9. Upload GPG public key to S3
10. Build and upload APT repository
11. Build and upload RPM/YUM repository
12. Upload installer scripts to S3
13. Purge CloudFlare CDN cache

### 3. Verify Workflow Completion

The workflow should complete within 30 minutes. If the GoReleaser step fails but artifacts were produced (e.g., Homebrew tap push failed due to permissions), subsequent steps still run due to `continue-on-error: true`.

Check workflow logs for:
- GoReleaser exit code and any partial failures
- S3 upload confirmation messages
- APT/RPM repository build output
- CloudFlare cache purge response

---

## Post-Release Verification

After the workflow completes, verify each distribution channel.

### GitHub Release

- [ ] Release exists at `github.com/StackEye-IO/stackeye-cli/releases/tag/vX.Y.Z`
- [ ] Changelog is populated with commit summaries
- [ ] All archives and packages are attached as assets

### CDN / Direct Download

```bash
VERSION=X.Y.Z  # without the v prefix

# Verify archive exists
curl -fsSI "https://releases.stackeye.io/cli/v${VERSION}/stackeye_${VERSION}_linux_amd64.tar.gz"

# Verify checksums
curl -fsSL "https://releases.stackeye.io/cli/v${VERSION}/checksums.txt"

# Verify GPG signature
curl -fsSL https://releases.stackeye.io/gpg-key.asc | gpg --import
curl -LO "https://releases.stackeye.io/cli/v${VERSION}/checksums.txt"
curl -LO "https://releases.stackeye.io/cli/v${VERSION}/checksums.txt.sig"
gpg --verify checksums.txt.sig checksums.txt
```

### Installer Scripts

```bash
# Linux/macOS installer
curl -fsSL https://releases.stackeye.io/install.sh | bash
stackeye version

# Windows (PowerShell)
iwr -useb https://releases.stackeye.io/install.ps1 | iex
stackeye version
```

### APT Repository (Debian/Ubuntu)

```bash
curl -fsSL https://releases.stackeye.io/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/stackeye-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/stackeye-archive-keyring.gpg] https://releases.stackeye.io stable main" | sudo tee /etc/apt/sources.list.d/stackeye.list > /dev/null
sudo apt-get update && sudo apt-get install stackeye
stackeye version
```

### YUM Repository (RHEL/Fedora/CentOS)

```bash
sudo rpm --import https://releases.stackeye.io/gpg-key.asc
sudo tee /etc/yum.repos.d/stackeye.repo << 'EOF'
[stackeye]
name=StackEye CLI
baseurl=https://releases.stackeye.io/yum/stable/$basearch
enabled=1
gpgcheck=1
gpgkey=https://releases.stackeye.io/gpg-key.asc
EOF
sudo dnf install stackeye
stackeye version
```

### Homebrew

```bash
brew update
brew install stackeye-io/tap/stackeye
stackeye version
```

Verify the formula was updated in [StackEye-IO/homebrew-tap](https://github.com/StackEye-IO/homebrew-tap).

### Scoop (Windows)

```powershell
scoop bucket add stackeye-io https://github.com/StackEye-IO/scoop-bucket
scoop install stackeye
stackeye version
```

Verify the manifest was updated in [StackEye-IO/scoop-bucket](https://github.com/StackEye-IO/scoop-bucket).

### Docker

```bash
# Latest tag
docker run --rm ghcr.io/stackeye-io/stackeye-cli:latest version

# Specific version
docker run --rm ghcr.io/stackeye-io/stackeye-cli:vX.Y.Z version

# Verify multi-arch manifest
docker manifest inspect ghcr.io/stackeye-io/stackeye-cli:vX.Y.Z
```

### Go Install

```bash
go install github.com/StackEye-IO/stackeye-cli/cmd/stackeye@vX.Y.Z
stackeye version
```

---

## Rollback Procedure

If a release has critical issues, follow these steps.

### 1. Delete the GitHub Release and Tag

```bash
# Delete the GitHub release (keeps the tag)
gh release delete vX.Y.Z --yes

# Delete the remote tag
git push origin --delete vX.Y.Z

# Delete the local tag
git tag -d vX.Y.Z
```

### 2. Remove CDN Artifacts

```bash
# Remove the version directory from S3
aws s3 rm "s3://releases.stackeye.io/cli/vX.Y.Z/" \
  --endpoint-url https://s3.us-central-1.wasabisys.com \
  --recursive

# Purge CDN cache
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"prefixes":["https://releases.stackeye.io/cli/vX.Y.Z/"]}'
```

### 3. Rebuild Package Repositories

The APT and YUM repositories only index the latest version. Re-tagging a previous good version and re-running the release workflow will overwrite them with correct packages.

```bash
# Re-tag and release a previous known-good version with a patch bump
git checkout v1.1.0  # last known-good tag
# Apply the critical fix
git tag -a v1.1.1 -m "Release v1.1.1 (rollback fix)"
git push origin v1.1.1
```

### 4. Revert Homebrew and Scoop

GoReleaser pushes formula/manifest updates automatically on the next release. For an immediate revert:

```bash
# Homebrew — revert the formula commit
cd /path/to/homebrew-tap
git revert HEAD
git push origin main

# Scoop — revert the manifest commit
cd /path/to/scoop-bucket
git revert HEAD
git push origin main
```

### 5. Remove Docker Images

```bash
# Delete specific tag from GHCR
gh api -X DELETE /orgs/StackEye-IO/packages/container/stackeye-cli/versions/<VERSION_ID>

# List versions to find the VERSION_ID
gh api /orgs/StackEye-IO/packages/container/stackeye-cli/versions
```

---

## Required Secrets

The release workflow requires these GitHub repository secrets:

| Secret | Purpose |
|--------|---------|
| `GITOPS_PAT` | Checkout private SDK, push Homebrew formula and Scoop manifest |
| `GPG_PRIVATE_KEY` | GPG signing of checksums and APT/RPM repositories |
| `GPG_PASSPHRASE` | GPG key passphrase (empty if key has no passphrase) |
| `WASABI_ACCESS_KEY_ID` | Wasabi S3 access for artifact uploads |
| `WASABI_SECRET_ACCESS_KEY` | Wasabi S3 secret for artifact uploads |
| `CLOUDFLARE_ZONE_ID` | CloudFlare zone for CDN cache purge |
| `CLOUDFLARE_API_TOKEN` | CloudFlare API token for cache purge |

`GITHUB_TOKEN` is provided automatically by GitHub Actions for GitHub Releases and GHCR.

---

## Versioning Policy

- Follow [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`
- **MAJOR**: Breaking changes to CLI flags, output format, or configuration
- **MINOR**: New commands, new flags, new features (backward compatible)
- **PATCH**: Bug fixes, documentation, dependency updates
- Pre-release tags (`-rc.N`, `-beta.N`) are marked as pre-release on GitHub automatically
- Tags must start with `v` (e.g., `v1.2.0`)
