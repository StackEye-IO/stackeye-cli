# Release Checklist

This document describes the end-to-end process for cutting a new StackEye CLI release. The release pipeline is fully automated via GitHub Actions — your job is to prepare the repository, push a tag, and verify the results.

## Distribution Channels

A single `git push --tags` triggers all of these:

| Channel | Destination | Mechanism | Published? |
|---------|-------------|-----------|------------|
| GitHub Releases | github.com/StackEye-IO/stackeye-cli/releases | GoReleaser | Yes |
| CDN Archives | releases.stackeye.io/cli/vX.Y.Z/ | S3 upload + CloudFlare | Yes |
| APT Repository | releases.stackeye.io/dists/stable/ | `scripts/build-apt-repo.sh` | Yes |
| YUM Repository | releases.stackeye.io/yum/stable/ | `scripts/build-rpm-repo.sh` | Yes |
| Homebrew Tap | StackEye-IO/homebrew-tap | GoReleaser brews | **No — see below** |
| Scoop Bucket | StackEye-IO/scoop-bucket | GoReleaser scoops | **No — see below** |
| Docker (GHCR) | ghcr.io/stackeye-io/stackeye-cli | GoReleaser dockers | Yes |
| Installer Scripts | releases.stackeye.io/install.sh, install.ps1 | S3 upload | Yes |
| GPG Signatures | releases.stackeye.io/gpg-key.asc | S3 upload | Yes |

> **Homebrew Tap and Scoop Bucket pushes 403 on every release — read before cutting one.**
>
> Every other channel above is confirmed live: verified 2026-08-23 against the `v0.2.0-rc.4`
> release (2026-08-20) with direct `curl` checks against `releases.stackeye.io` — CDN
> archives, `gpg-key.asc`/`checksums.txt.sig`, and both the APT and YUM repositories
> (`dists/stable/`, `yum/stable/<arch>/`) all resolve. The R2 migration (stackeye-6087) and
> the release-tooling/hard-fail fixes from stackeye-6095 and stackeye-6292 are done; this
> doc's earlier "still points at Wasabi" / "never run" blockers no longer apply.
>
> What is still broken: the `GITOPS_PAT` fetched from Vault (see step 3 below) lacks write
> access to `StackEye-IO/homebrew-tap` and `StackEye-IO/scoop-bucket`, so GoReleaser's brew
> and scoop pushes fail with `403 Resource not accessible by personal access token` on every
> run. The GoReleaser step's `continue-on-error: true` exists specifically to tolerate this,
> so the release itself still succeeds — but `Formula/stackeye.rb` and `stackeye.json` are
> stuck at stale versions (verified 2026-08-23: the tap is still at `v0.1.0-alpha.2`, the
> bucket at `v0.2.0-rc.2`, while the latest published release is `v0.2.0-rc.4`). Tracked as
> stackeye-6236 (task 21418, `backlog` as of this writing). Do not tell customers
> `brew install`/`scoop install` will get the latest release until that lands.

---

## Pre-Release Checklist

Complete these items before tagging.

- [ ] **All CI checks pass on `main`** — verify at [Actions](https://github.com/StackEye-IO/stackeye-cli/actions/workflows/ci.yml)
- [ ] **Run full local validation**:
  ```bash
  make validate
  ```
- [ ] **CHANGELOG updated** (optional) — if `CHANGELOG.md` exists, add a section for the new version. GoReleaser auto-generates release notes from conventional commit messages, so a curated changelog is only needed for major releases
- [ ] **Version string is correct** — the version comes from the git tag, not from source. Verify `internal/version` uses ldflags injection (no hardcoded version)
- [ ] **Dependencies are current** — check for any pending Dependabot PRs that should be merged first
- [ ] **No unmerged breaking changes** — confirm all intended PRs are merged to `main`
- [ ] **Command reference is current** — if any command, flag, or example changed this cycle, regenerate the docs and reconcile the hand-written pages (see [Command Documentation](#command-documentation) below)
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
2. Install release tooling (`aws-cli`, `gnupg2`) — the `self-hosted-linux-dfw` runner image
   doesn't ship these, so they're bootstrapped in-job (stackeye-6095)
3. Fetch `GITOPS_PAT` from Vault (keyless OIDC, `secret/data/stackeye/ci/gitops-pat`)
4. Checkout `stackeye-go-sdk` (private dependency, using the fetched `GITOPS_PAT`)
5. Setup Go
6. Update the `go.mod` replace directive to the checked-out SDK path (`./.sdk`) and `go mod tidy`
7. Download dependencies
8. Import GPG signing key
9. Setup Docker Buildx + QEMU (multi-arch)
10. Login to GHCR
11. **GoReleaser** — builds binaries, archives, .deb/.rpm, Docker images, GitHub Release, Homebrew formula, Scoop manifest, GPG-signed checksums (`continue-on-error: true`, to tolerate the known Homebrew/Scoop 403 — see the blocker above)
12. Assert GoReleaser published a release — hard-fails the job if `dist/checksums.txt` is missing or the GitHub release has zero assets, so a silently-skipped publish can't pass as green (stackeye-6292)
13. Upload artifacts to Cloudflare R2 (bucket `stackeye-releases`)
14. Upload GPG public key to R2
15. Build and upload APT repository
16. Build and upload RPM/YUM repository
17. Upload installer scripts to R2
18. Purge CloudFlare CDN cache

All S3-compatible steps (13-17) target the Cloudflare R2 bucket `stackeye-releases` via
`secrets.R2_S3_ENDPOINT` — there is no remaining Wasabi dependency anywhere in this workflow.

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
```

Verify with:

```bash
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

Verify on a clean container:

```bash
curl -fsSL https://releases.stackeye.io/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/stackeye-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/stackeye-archive-keyring.gpg] https://releases.stackeye.io stable main" | sudo tee /etc/apt/sources.list.d/stackeye.list > /dev/null
sudo apt-get update && sudo apt-get install stackeye
stackeye version
```

### YUM Repository (RHEL/Fedora/CentOS)

Verify on a clean container:

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
This currently fails on every release — see the blocker under
[Distribution Channels](#distribution-channels) (stackeye-6236 / task 21418).

### Scoop (Windows)

```powershell
scoop bucket add stackeye-io https://github.com/StackEye-IO/scoop-bucket
scoop install stackeye
stackeye version
```

Verify the manifest was updated in [StackEye-IO/scoop-bucket](https://github.com/StackEye-IO/scoop-bucket).
This currently fails on every release — see the blocker under
[Distribution Channels](#distribution-channels) (stackeye-6236 / task 21418).

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

Requires access to the `StackEye-IO` GitHub organization (the SDK dependency is in a private repository).

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
# Remove the version directory from R2 (endpoint is the R2_S3_ENDPOINT secret value)
aws s3 rm "s3://stackeye-releases/cli/vX.Y.Z/" \
  --endpoint-url "$R2_S3_ENDPOINT" \
  --recursive

# Purge CDN cache. NOTE: the "prefixes" purge type takes bare host+path, NOT
# full URLs — a leading https:// makes Cloudflare reject the whole request
# (error 1119, stackeye-6096).
curl --fail -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"prefixes":["releases.stackeye.io/cli/vX.Y.Z/"]}'
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
| `GPG_PRIVATE_KEY` | GPG signing of checksums and APT/RPM repositories |
| `GPG_PASSPHRASE` | GPG key passphrase (empty if key has no passphrase) |
| `R2_ACCESS_KEY_ID` | Cloudflare R2 S3-compatible access key for artifact uploads |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 S3-compatible secret key for artifact uploads |
| `R2_S3_ENDPOINT` | Cloudflare R2 S3-compatible endpoint URL for the `stackeye-releases` bucket |
| `CLOUDFLARE_ZONE_ID` | CloudFlare zone for CDN cache purge |
| `CLOUDFLARE_API_TOKEN` | CloudFlare API token for cache purge |

`GITOPS_PAT` (checkout of the private SDK, and the Homebrew/Scoop pushes — see the blocker
under [Distribution Channels](#distribution-channels)) is **not** a static GitHub secret; it's
fetched at job runtime from Vault via keyless OIDC (`secret/data/stackeye/ci/gitops-pat`).

`GITHUB_TOKEN` is provided automatically by GitHub Actions for GitHub Releases and GHCR.

---

## Command Documentation

The CLI command reference is generated directly from the cobra command tree, so it
always matches the implemented commands, flags, and short descriptions.

### Generation targets

| Target | Output | Format |
|--------|--------|--------|
| `make markdown` | `docs/markdown/pages/` | One Markdown page per command (via `cobra/doc`) |
| `make man` | `docs/man/pages/` | roff man pages |

Both generators (`docs/markdown/generate.go`, `docs/man/generate.go`) build the root
command with `PersistentPreRunE` disabled, so they run without a real config file.

### Generated pages are ephemeral

`docs/markdown/pages/` and `docs/man/pages/` are **git-ignored** (see `.gitignore`) and
are **not** committed. They are regenerated on demand and consumed by the docs site /
man-page packaging. Do **not** hand-edit them — changes are overwritten on the next run.
Keep all hand-written guidance in the source-of-truth docs instead:

- `README.md` — command tables, installation, quick start, exit codes
- `docs/getting-started.md` — end-to-end tutorial
- `examples/` — runnable example configs and recipes

### SDK replace directive

`go.mod` has `replace github.com/StackEye-IO/stackeye-go-sdk => ../stackeye-go-sdk`.
Generation (and any local build) therefore needs the `stackeye-go-sdk` checkout beside
this repo. CI rewrites the directive to a local `./.sdk` checkout; locally, either check
the SDK out at `../stackeye-go-sdk` or point the replace directive at your copy.

### Regenerate & reconcile before a release

```bash
# 1. Regenerate the command reference from current code
make markdown        # and/or: make man

# 2. Diff generated pages against the hand-written docs and reconcile any drift:
#    - new/renamed commands or flags -> update README.md command tables
#    - changed example syntax        -> update docs/getting-started.md and examples/
#    Use template variables ({probe_id}, {api_key}, ...) in examples — never fake data.

# 3. Clean up generated pages when done (they are not committed)
make markdown-clean  # and/or: make man-clean
```

## Versioning Policy

- Follow [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`
- **MAJOR**: Breaking changes to CLI flags, output format, or configuration
- **MINOR**: New commands, new flags, new features (backward compatible)
- **PATCH**: Bug fixes, documentation, dependency updates
- Pre-release tags (`-rc.N`, `-beta.N`) are marked as pre-release on GitHub automatically
- Tags must start with `v` (e.g., `v1.2.0`)
