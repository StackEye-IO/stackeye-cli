# StackEye CLI Examples

This directory contains example configurations and recipes for the StackEye CLI.

## Directory Structure

```
examples/
├── probes/          # Probe configuration examples
├── channels/        # Notification channel configurations
├── status-pages/    # Status page configurations
├── cicd/            # CI/CD integration examples
└── scripts/         # Helper scripts
```

## Quick Start

### 1. Set Up Authentication

```bash
# Interactive login (recommended for development)
stackeye login

# Or use API key for automation
export STACKEYE_API_KEY=se_{your_api_key}
stackeye setup --no-input
```

### 2. Create Resources from Examples

```bash
# Create a notification channel from YAML
stackeye channel create --from-file examples/channels/slack.yaml

# Create a status page from YAML
stackeye status-page create --from-file examples/status-pages/basic.yaml

# Create a probe (using command-line flags)
stackeye probe create \
  --name "Production API" \
  --url https://api.example.com/health \
  --interval 60 \
  --timeout 10
```

### 3. Link Channels to Probes

```bash
# List your probes and channels to get IDs
stackeye probe list
stackeye channel list

# Link a channel to receive alerts for a probe
stackeye probe link-channel "{probe_id}" "{channel_id}"
```

## Configuration Patterns

### GitOps Workflow

Store your monitoring configuration in version control:

```bash
# Directory structure
monitoring/
├── channels/
│   ├── production-slack.yaml
│   └── oncall-pagerduty.yaml
├── status-pages/
│   └── public-status.yaml
└── scripts/
    └── apply-config.sh
```

Apply configurations during deployment:

```bash
#!/bin/bash
# apply-config.sh

# Apply all channel configurations
for f in monitoring/channels/*.yaml; do
  stackeye channel create --from-file "$f" || true
done

# Apply status page configurations
for f in monitoring/status-pages/*.yaml; do
  stackeye status-page create --from-file "$f" || true
done
```

### Environment-Specific Configurations

Use environment variables in your CI/CD pipeline:

```yaml
# channels/slack-template.yaml
name: "${ENV}-alerts"
type: slack
config:
  webhook_url: "${SLACK_WEBHOOK_URL}"
```

Then substitute at deploy time:

```bash
envsubst < channels/slack-template.yaml | stackeye channel create --from-file -
```

## Example Categories

### Probes (`probes/`)

| File | Description |
|------|-------------|
| `http-basic.yaml` | Simple HTTP health check |
| `http-advanced.yaml` | HTTP probe with all options |
| `tcp-probe.yaml` | TCP port connectivity check |
| `dns-probe.yaml` | DNS resolution check |
| `multi-probe.yaml` | Multiple probes in one script |

**Note**: Probe creation currently uses command-line flags. The YAML files in `probes/` show the equivalent configuration as documentation.

### Channels (`channels/`)

| File | Description |
|------|-------------|
| `email.yaml` | Email notifications |
| `slack.yaml` | Slack webhook integration |
| `discord.yaml` | Discord webhook integration |
| `teams.yaml` | Microsoft Teams integration |
| `pagerduty.yaml` | PagerDuty incident creation |
| `webhook.yaml` | Generic webhook (custom integrations) |
| `sms.yaml` | SMS notifications |

### Status Pages (`status-pages/`)

| File | Description |
|------|-------------|
| `basic.yaml` | Minimal status page configuration |
| `branded.yaml` | Fully customized with branding |

### CI/CD (`cicd/`)

| File | Description |
|------|-------------|
| `github-actions.yaml` | GitHub Actions workflow |
| `gitlab-ci.yaml` | GitLab CI pipeline |
| `jenkins.groovy` | Jenkins pipeline script |

### Scripts (`scripts/`)

| File | Description |
|------|-------------|
| `bootstrap.sh` | One-command initial setup |
| `monitor-deployment.sh` | Post-deployment monitoring setup |

## Best Practices

### 1. Use Descriptive Names

```bash
# Good - clear purpose
stackeye probe create --name "Production API - Health Check" ...

# Avoid - ambiguous
stackeye probe create --name "api" ...
```

### 2. Set Appropriate Intervals

| Use Case | Recommended Interval |
|----------|---------------------|
| Critical production services | 30-60 seconds |
| Standard services | 60-120 seconds |
| Non-critical/batch jobs | 300+ seconds |

### 3. Configure Multiple Notification Channels

```bash
# Create separate channels for different severity levels
stackeye channel create --name "Slack - Warnings" --type slack ...
stackeye channel create --name "PagerDuty - Critical" --type pagerduty ...

# Link appropriate channels to probes
stackeye probe link-channel "critical-api" "{pagerduty_channel_id}"
stackeye probe link-channel "critical-api" "{slack_channel_id}"
```

### 4. Use Labels for Organization

```bash
# Add labels for filtering and organization
stackeye probe label "{probe_id}" env=production tier=api team=backend

# List probes by label
stackeye probe list --labels "env=production"
```

## Troubleshooting

### Authentication Issues

```bash
# Check current authentication status
stackeye whoami

# View full configuration
stackeye config get

# Re-authenticate
stackeye logout
stackeye login
```

### YAML Parsing Errors

```bash
# Validate YAML syntax before applying
python3 -c "import yaml; yaml.safe_load(open('channel.yaml'))"

# Or use yq
yq '.' channel.yaml
```

### Debug Mode

```bash
# Enable verbose logging for troubleshooting
stackeye --debug channel create --from-file channel.yaml
```

## Getting Help

- **Command Help**: `stackeye <command> --help`
- **Documentation**: [docs.stackeye.io](https://docs.stackeye.io)
- **Getting Started Tutorial**: [docs/getting-started.md](../docs/getting-started.md)
- **Support**: [GitHub Issues](https://github.com/StackEye-IO/stackeye-cli/issues)
