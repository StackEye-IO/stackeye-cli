# StackEye CLI

[![License](https://img.shields.io/github/license/StackEye-IO/stackeye-cli)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/StackEye-IO/stackeye-cli)](https://goreportcard.com/report/github.com/StackEye-IO/stackeye-cli)
[![Release](https://img.shields.io/github/v/release/StackEye-IO/stackeye-cli)](https://github.com/StackEye-IO/stackeye-cli/releases/latest)
[![CI](https://github.com/StackEye-IO/stackeye-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/StackEye-IO/stackeye-cli/actions/workflows/ci.yml)

Command-line interface for [StackEye](https://stackeye.io) - the full-stack uptime monitoring platform.

## Overview

The StackEye CLI (`stackeye`) provides command-line access to the StackEye uptime monitoring platform. Authenticate with your account, manage configuration contexts, and integrate monitoring into your automation workflows.

**Current Features:**
- Browser-based authentication via OAuth
- Multi-context configuration management
- API key management
- Probe management (list, create, update, delete, pause, resume)
- Alert management (list, acknowledge, resolve)
- Notification channel management
- Organization switching
- Dashboard view
- Shell completion for bash, zsh, fish, and PowerShell

## Installation

### Installer Script (Recommended for macOS/Linux)

The easiest way to install StackEye CLI:

```bash
curl -fsSL https://get.stackeye.io/cli | bash
```

This auto-detects your OS and architecture, downloads the appropriate binary, verifies the checksum, and installs it.

**Options:**
- Install a specific version: `curl -fsSL https://get.stackeye.io/cli | bash -s -- --version v1.0.0`
- Install to custom directory: `STACKEYE_INSTALL_DIR=~/bin curl -fsSL https://get.stackeye.io/cli | bash`

The script installs to `/usr/local/bin` (with sudo) or `~/.local/bin` (without sudo) by default.

### Manual Download

Download pre-built archives from [GitHub Releases](https://github.com/StackEye-IO/stackeye-cli/releases/latest). Archives are named `stackeye_<VERSION>_<OS>_<ARCH>.tar.gz`.

**macOS (Apple Silicon)**:
```bash
# Replace VERSION with the actual version (e.g., 1.0.0)
VERSION=$(curl -fsSL https://api.github.com/repos/StackEye-IO/stackeye-cli/releases/latest | grep tag_name | sed 's/.*"v\([^"]*\)".*/\1/')
curl -Lo stackeye.tar.gz "https://github.com/StackEye-IO/stackeye-cli/releases/latest/download/stackeye_${VERSION}_darwin_arm64.tar.gz"
tar -xzf stackeye.tar.gz
sudo mv stackeye /usr/local/bin/
rm stackeye.tar.gz
```

**macOS (Intel)**:
```bash
VERSION=$(curl -fsSL https://api.github.com/repos/StackEye-IO/stackeye-cli/releases/latest | grep tag_name | sed 's/.*"v\([^"]*\)".*/\1/')
curl -Lo stackeye.tar.gz "https://github.com/StackEye-IO/stackeye-cli/releases/latest/download/stackeye_${VERSION}_darwin_amd64.tar.gz"
tar -xzf stackeye.tar.gz
sudo mv stackeye /usr/local/bin/
rm stackeye.tar.gz
```

**Linux (x86_64)**:
```bash
VERSION=$(curl -fsSL https://api.github.com/repos/StackEye-IO/stackeye-cli/releases/latest | grep tag_name | sed 's/.*"v\([^"]*\)".*/\1/')
curl -Lo stackeye.tar.gz "https://github.com/StackEye-IO/stackeye-cli/releases/latest/download/stackeye_${VERSION}_linux_amd64.tar.gz"
tar -xzf stackeye.tar.gz
sudo mv stackeye /usr/local/bin/
rm stackeye.tar.gz
```

**Linux (ARM64)**:
```bash
VERSION=$(curl -fsSL https://api.github.com/repos/StackEye-IO/stackeye-cli/releases/latest | grep tag_name | sed 's/.*"v\([^"]*\)".*/\1/')
curl -Lo stackeye.tar.gz "https://github.com/StackEye-IO/stackeye-cli/releases/latest/download/stackeye_${VERSION}_linux_arm64.tar.gz"
tar -xzf stackeye.tar.gz
sudo mv stackeye /usr/local/bin/
rm stackeye.tar.gz
```

**Windows (PowerShell)**:
```powershell
# Get latest version
$VERSION = (Invoke-RestMethod "https://api.github.com/repos/StackEye-IO/stackeye-cli/releases/latest").tag_name.TrimStart('v')

# Download and extract
Invoke-WebRequest -Uri "https://github.com/StackEye-IO/stackeye-cli/releases/latest/download/stackeye_${VERSION}_windows_amd64.zip" -OutFile stackeye.zip
Expand-Archive stackeye.zip -DestinationPath .
Move-Item stackeye.exe C:\Windows\System32\stackeye.exe
Remove-Item stackeye.zip
```

### Go Install

If you have Go 1.25+ installed:

```bash
go install github.com/StackEye-IO/stackeye-cli/cmd/stackeye@latest
```

The binary will be installed to `$GOPATH/bin/stackeye` (or `$HOME/go/bin/stackeye` if `GOPATH` is not set).

### Build from Source

Requires Go 1.25 or later:

```bash
# Clone the repository
git clone https://github.com/StackEye-IO/stackeye-cli.git
cd stackeye-cli

# Build the binary
make build

# The binary is at bin/stackeye
# Optionally move to your PATH:
sudo mv bin/stackeye /usr/local/bin/
```

### Homebrew (macOS/Linux)

```bash
brew install stackeye-io/tap/stackeye
```

### Scoop (Windows)

```powershell
# Add the StackEye bucket
scoop bucket add stackeye-io https://github.com/StackEye-IO/scoop-bucket

# Install
scoop install stackeye

# Update
scoop update stackeye
```

### Coming Soon

The following installation methods will be available in future releases:

| Method | Platform | Status |
|--------|----------|--------|
| APT | Debian/Ubuntu | `.deb` packages |
| RPM | RHEL/Fedora | `.rpm` packages |
| Docker | All | `docker run ghcr.io/stackeye-io/stackeye-cli` |

## Quick Start

### 1. Authenticate

```bash
# Login via browser (opens your default browser)
stackeye login

# Output:
# Opening browser to: https://stackeye.io/cli-auth?...
# Waiting for authentication...
# (If the browser doesn't open, visit the URL manually)
#
# Verifying credentials... done
#
# Successfully logged in!
#   Organization: {org_name}
#   Context:      {context_name}
#   API URL:      https://api.stackeye.io
#
# Credentials saved to: ~/.config/stackeye/config.yaml
```

### 2. Verify Authentication

```bash
# Check current user
stackeye whoami

# Output:
#
# User:         {user_email}
# Name:         {user_name}
#
# Context:      {context_name}
# Organization: {org_name} ({org_id})
# API URL:      https://api.stackeye.io
# Auth Type:    api_key
```

### 3. View Configuration

```bash
# Show current configuration
stackeye config get

# Output:
# Current Context:    {context_name}
# API URL:            https://api.stackeye.io
# API Key:            se_****...xxxx
# Organization:       {org_name}
# Organization ID:    {org_id}
# Config File:        ~/.config/stackeye/config.yaml
```

## CI/CD and Automation

For non-interactive environments (CI/CD pipelines, scripts, Docker containers), the CLI supports direct API key authentication.

### Environment Variable (Recommended)

```bash
# Set API key via environment variable
export STACKEYE_API_KEY=se_your_api_key_here

# Setup and verify configuration
stackeye setup --no-input

# Create probes using stored credentials
stackeye probe create --name "Production API" --url https://api.example.com
```

### Command Flags (Single-Command Setup)

```bash
# Full setup with API key and probe creation
stackeye setup --no-input \
  --api-key se_your_api_key_here \
  --probe-name "Production API" \
  --probe-url https://api.example.com

# API key only (no probe creation)
stackeye setup --no-input --api-key se_your_api_key_here
```

### GitHub Actions Example

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Setup StackEye monitoring
        env:
          STACKEYE_API_KEY: ${{ secrets.STACKEYE_API_KEY }}
        run: |
          curl -fsSL https://get.stackeye.io/cli | bash
          stackeye setup --no-input
          stackeye probe create --name "${{ github.repository }}" --url "https://api.example.com"
```

### Configuration Precedence

1. `--api-key` command flag (highest)
2. `STACKEYE_API_KEY` environment variable
3. Config file (`~/.config/stackeye/config.yaml`)

## Commands

### Authentication & Configuration

| Command | Description |
|---------|-------------|
| `stackeye login` | Authenticate with StackEye via browser |
| `stackeye logout` | Clear stored credentials for current context |
| `stackeye setup` | Interactive setup wizard for first-time users |
| `stackeye whoami` | Display current authenticated user |
| `stackeye config get` | Display current configuration |
| `stackeye config set-key` | Set API key for authentication |
| `stackeye context list` | List all configured contexts |
| `stackeye context use <name>` | Switch to a different context |
| `stackeye context current` | Display active context |
| `stackeye completion <shell>` | Generate shell completion script |
| `stackeye version` | Print version information |

### Probe Management

| Command | Description |
|---------|-------------|
| `stackeye probe list` | List all monitoring probes |
| `stackeye probe get <id>` | Get probe details |
| `stackeye probe create` | Create a new probe |
| `stackeye probe update <id>` | Update probe configuration |
| `stackeye probe delete <id>` | Delete a probe |
| `stackeye probe pause <id>` | Pause probe monitoring |
| `stackeye probe resume <id>` | Resume probe monitoring |
| `stackeye probe test <id>` | Run an immediate probe check |
| `stackeye probe history <id>` | View probe check history |
| `stackeye probe stats <id>` | View probe statistics |

### Alert Management

| Command | Description |
|---------|-------------|
| `stackeye alert list` | List current alerts |
| `stackeye alert get <id>` | Get alert details |
| `stackeye alert ack <id>` | Acknowledge an alert |
| `stackeye alert resolve <id>` | Resolve an alert |
| `stackeye alert history` | View alert history |

### Notification Channels

| Command | Description |
|---------|-------------|
| `stackeye channel list` | List notification channels |
| `stackeye channel get <id>` | Get channel details |
| `stackeye channel create` | Create a notification channel |
| `stackeye channel update <id>` | Update channel configuration |
| `stackeye channel delete <id>` | Delete a channel |
| `stackeye channel test <id>` | Send a test notification |

### Organization & Dashboard

| Command | Description |
|---------|-------------|
| `stackeye org list` | List accessible organizations |
| `stackeye org get` | Get current organization details |
| `stackeye org switch <id>` | Switch to a different organization |
| `stackeye dashboard` | Display dashboard overview |
| `stackeye region list` | List available monitoring regions |
| `stackeye apikey list` | List API keys |
| `stackeye apikey create` | Create a new API key |

Run `stackeye --help` for complete command documentation.

### Authentication Commands

#### `stackeye login`

Authenticate with StackEye via browser-based OAuth flow:

```bash
# Login to production StackEye
stackeye login

# Login to a specific environment
stackeye login --api-url https://api.dev.stackeye.io
```

#### `stackeye logout`

Clear stored credentials:

```bash
# Logout from current context
stackeye logout

# Logout from all contexts
stackeye logout --all
```

#### `stackeye whoami`

Display information about the authenticated user:

```bash
stackeye whoami
```

### Configuration Commands

#### `stackeye config get`

Display current configuration:

```bash
stackeye config get
```

#### `stackeye config set-key`

Set an API key directly (alternative to browser login):

```bash
# Set API key interactively (recommended)
stackeye config set-key

# Set API key with verification
stackeye config set-key --verify

# Set API key for specific context
stackeye config set-key --context production
```

### Context Commands

Contexts allow switching between different organizations or environments, similar to kubectl contexts for Kubernetes.

#### `stackeye context list`

List all configured contexts:

```bash
stackeye context list

# Output:
#    NAME                 ORGANIZATION              API URL
# *  {context_name}       {org_name}                https://api.stackeye.io
#    {context_name_dev}   {org_name}                https://api.dev.stackeye.io
```

#### `stackeye context use`

Switch to a different context:

```bash
stackeye context use {context_name}
```

#### `stackeye context current`

Display the active context:

```bash
stackeye context current
```

## Configuration

### Config File Location

Configuration is stored in `~/.config/stackeye/config.yaml`:

```yaml
# Current active context
current_context: {context_name}

# Configured contexts
contexts:
  {context_name}:
    api_url: https://api.stackeye.io
    api_key: se_{api_key_suffix}
    organization_id: {org_id}
    organization_name: {org_name}
  {context_name_dev}:
    api_url: https://api.dev.stackeye.io
    api_key: se_{api_key_suffix_dev}
    organization_name: {org_name}

# User preferences
preferences:
  output_format: table    # table, json, yaml, wide
  color: auto             # auto, always, never
  debug: false
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `STACKEYE_API_URL` | Override API URL | `https://api.stackeye.io` |
| `STACKEYE_API_KEY` | Override API key | (from config) |
| `STACKEYE_CONFIG` | Custom config file path | `~/.config/stackeye/config.yaml` |
| `NO_COLOR` | Disable colored output | (unset) |

## Global Flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `--config <path>` | Use custom config file |
| `--context <name>` | Override current context |
| `--output, -o <format>` | Output format: table, json, yaml, wide |
| `--no-color` | Disable colored output |
| `--no-input` | Disable interactive prompts |
| `--dry-run` | Show what would be done |
| `--debug` | Enable debug output |
| `--help, -h` | Show help |

## Shell Completion

Enable tab completion for your shell:

### Bash

```bash
# Add to ~/.bashrc
source <(stackeye completion bash)

# Or install to system-wide location
stackeye completion bash > /etc/bash_completion.d/stackeye
```

### Zsh

```bash
# Add to ~/.zshrc (before compinit)
source <(stackeye completion zsh)

# Or install to fpath
stackeye completion zsh > "${fpath[1]}/_stackeye"
```

### Fish

```fish
# Add to ~/.config/fish/config.fish
stackeye completion fish | source

# Or install to completions directory
stackeye completion fish > ~/.config/fish/completions/stackeye.fish
```

### PowerShell

```powershell
# Add to your PowerShell profile
stackeye completion powershell | Out-String | Invoke-Expression

# Or save to a file and dot-source it
stackeye completion powershell > stackeye.ps1
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Command misuse (invalid arguments) |
| 3 | Authentication required |
| 4 | Permission denied |
| 5 | Resource not found |
| 6 | Rate limited |
| 7 | Server error |
| 8 | Network error |
| 9 | Timeout |
| 10 | Plan limit exceeded |

## Roadmap

Planned features for upcoming releases:

- **Status Pages**: `stackeye status list`, `create`, `update`
- **Watch Mode**: `stackeye watch` for live terminal updates
- **Incident Management**: `stackeye incident list`, `create`, `update`
- **Team Management**: `stackeye team list`, `invite`, `remove`
- **Maintenance Windows**: `stackeye maintenance list`, `create`

## Contributing

Contributions are welcome!

### Development Setup

```bash
# Clone the repository
git clone https://github.com/StackEye-IO/stackeye-cli.git
cd stackeye-cli

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linters
make lint

# Full validation
make validate
```

Submit issues and pull requests on [GitHub](https://github.com/StackEye-IO/stackeye-cli).

## Support

- **Documentation**: [docs.stackeye.io](https://docs.stackeye.io)
- **Issues**: [GitHub Issues](https://github.com/StackEye-IO/stackeye-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/StackEye-IO/stackeye-cli/discussions)

## License

MIT License - see [LICENSE](LICENSE) for details.
