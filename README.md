# StackEye CLI

Command-line interface for [StackEye](https://stackeye.io) - the full-stack uptime monitoring platform.

## Overview

The StackEye CLI (`stackeye`) provides a powerful command-line interface for managing your uptime monitoring infrastructure. Create probes, configure alerts, manage notification channels, and view your monitoring status - all from the terminal.

## Installation

### Homebrew (macOS)

```bash
brew install stackeye-io/tap/stackeye
```

### Scoop (Windows)

```bash
scoop bucket add stackeye https://github.com/StackEye-IO/scoop-bucket.git
scoop install stackeye
```

### Go Install

```bash
go install github.com/StackEye-IO/stackeye-cli@latest
```

### Binary Download

Download pre-built binaries from the [Releases](https://github.com/StackEye-IO/stackeye-cli/releases) page.

## Quick Start

```bash
# Authenticate with your StackEye account
stackeye login

# List your probes
stackeye probe list

# Create a new HTTP probe
stackeye probe create --name "API Health" --url https://api.example.com/health

# View dashboard summary
stackeye dashboard
```

## Commands

| Command | Description |
|---------|-------------|
| `stackeye login` | Authenticate with StackEye |
| `stackeye logout` | Clear authentication |
| `stackeye probe list` | List all probes |
| `stackeye probe create` | Create a new probe |
| `stackeye alert list` | List active alerts |
| `stackeye channel list` | List notification channels |
| `stackeye org list` | List organizations |
| `stackeye status list` | List status pages |

Run `stackeye --help` for a complete list of commands.

## Configuration

The CLI stores configuration in `~/.config/stackeye/config.yaml`:

```yaml
current_context: acme-prod
contexts:
  acme-prod:
    api_url: https://api.stackeye.io
    api_key: se_...
```

## Documentation

- [CLI Documentation](https://docs.stackeye.io/cli)
- [API Reference](https://docs.stackeye.io/api)
- [Getting Started Guide](https://docs.stackeye.io/getting-started)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
