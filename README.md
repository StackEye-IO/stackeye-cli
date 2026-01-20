# StackEye CLI

Command-line interface for [StackEye](https://stackeye.io) - the full-stack uptime monitoring platform.

## Overview

The StackEye CLI (`stackeye`) provides command-line access to the StackEye uptime monitoring platform. Authenticate with your account, manage configuration contexts, and integrate monitoring into your automation workflows.

**Current Features:**
- Browser-based authentication via OAuth
- Multi-context configuration management
- API key management
- Shell completion for bash, zsh, fish, and PowerShell

## Installation

### Build from Source

Requires Go 1.22 or later:

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

### Coming Soon

The following installation methods will be available in future releases:

- **Homebrew (macOS)**: `brew install stackeye-io/tap/stackeye`
- **Scoop (Windows)**: `scoop install stackeye`
- **Binary Downloads**: Pre-built binaries for macOS, Linux, and Windows
- **Go Install**: `go install github.com/StackEye-IO/stackeye-cli/cmd/stackeye@latest`

## Quick Start

### 1. Authenticate

```bash
# Login via browser (opens your default browser)
stackeye login

# Output:
# Opening browser to: https://stackeye.io/cli-auth?...
# Waiting for authentication...
# Verifying credentials... done
#
# Successfully logged in!
#   User:         John Doe (john@example.com)
#   Organization: Example Corp
#   Context:      example-corp
#   API URL:      https://api.stackeye.io
```

### 2. Verify Authentication

```bash
# Check current user
stackeye whoami

# Output:
# User: John Doe (john@example.com)
# Organization: Example Corp (org_...)
# Role: Admin
```

### 3. View Configuration

```bash
# Show current configuration
stackeye config get

# Output:
# Current Context:    example-corp
# API URL:            https://api.stackeye.io
# API Key:            se_****...xxxx
# Organization:       Example Corp
# Config File:        /home/user/.config/stackeye/config.yaml
```

## Commands

| Command | Description |
|---------|-------------|
| `stackeye login` | Authenticate with StackEye via browser |
| `stackeye logout` | Clear stored credentials for current context |
| `stackeye whoami` | Display current authenticated user |
| `stackeye config get` | Display current configuration |
| `stackeye config set-key` | Set API key for authentication |
| `stackeye context list` | List all configured contexts |
| `stackeye context use <name>` | Switch to a different context |
| `stackeye context current` | Display active context |
| `stackeye completion <shell>` | Generate shell completion script |
| `stackeye version` | Print version information |

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
# *  example-corp         Example Corp              https://api.stackeye.io
#    example-corp-dev     Example Corp              https://api.dev.stackeye.io
```

#### `stackeye context use`

Switch to a different context:

```bash
stackeye context use example-corp-dev
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
current_context: example-corp

# Configured contexts
contexts:
  example-corp:
    api_url: https://api.stackeye.io
    api_key: se_abc123...
    organization_id: org_...
    organization_name: Example Corp
  example-corp-dev:
    api_url: https://api.dev.stackeye.io
    api_key: se_def456...
    organization_name: Example Corp

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

- **Probe Management**: `stackeye probe list`, `create`, `update`, `delete`, `pause`, `resume`
- **Alert Management**: `stackeye alert list`, `ack`, `resolve`
- **Channel Management**: `stackeye channel list`, `create`, `test`
- **Organization Switching**: `stackeye org list`, `switch`
- **Status Pages**: `stackeye status list`, `create`
- **Dashboard View**: `stackeye dashboard`
- **Watch Mode**: `stackeye watch` for live updates

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

## License

MIT License - see [LICENSE](LICENSE) for details.
