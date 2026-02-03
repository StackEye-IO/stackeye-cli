# Getting Started with StackEye CLI

This tutorial walks you through setting up monitoring for your first endpoint using the StackEye CLI. By the end, you'll have a working probe that monitors your service and sends alerts when issues are detected.

**Time to complete**: 10-15 minutes

## Prerequisites

Before you begin, you'll need:

- A StackEye account ([sign up at stackeye.io](https://stackeye.io))
- A URL or endpoint to monitor (your website, API, or any HTTP endpoint)
- Terminal access (macOS, Linux, or Windows PowerShell)

## Step 1: Install the CLI

### Quick Install (macOS/Linux)

```bash
curl -fsSL https://get.stackeye.io/cli | bash
```

### Other Installation Methods

See the [README](../README.md#installation) for additional installation options including:
- Homebrew (macOS/Linux)
- Scoop (Windows)
- Manual download
- Build from source

### Verify Installation

```bash
stackeye version
```

You should see output like:

```
StackEye CLI
  Version:    {version}
  Commit:     {commit}
  Built:      {build_date}
  Go version: {go_version}
  OS/Arch:    {os}/{arch}
```

## Step 2: Authenticate

The CLI uses browser-based authentication to securely connect to your StackEye account.

```bash
stackeye login
```

This will:
1. Open your default browser to the StackEye login page
2. Ask you to sign in (or use your existing session)
3. Authorize the CLI to access your account
4. Save credentials locally

**Expected output:**

```
Opening browser to: https://stackeye.io/cli-auth?...
Waiting for authentication...

Verifying credentials... done

Successfully logged in!
  Organization: {your_org_name}
  Context:      {your_org_name}
  API URL:      https://api.stackeye.io

Credentials saved to: ~/.config/stackeye/config.yaml
```

### Verify Authentication

```bash
stackeye whoami
```

**Expected output:**

```
User:         {your_email}
Name:         {your_name}

Context:      {your_org_name}
Organization: {your_org_name} ({org_id})
API URL:      https://api.stackeye.io
Auth Type:    api_key
```

## Step 3: Create Your First Probe

A **probe** is a monitoring check that runs at regular intervals from multiple global regions. Let's create one to monitor an HTTP endpoint.

### Basic HTTP Probe

```bash
stackeye probe create \
  --name "My Website" \
  --url https://example.com
```

**Expected output:**

```
Probe created successfully!

ID:           {probe_id}
Name:         My Website
URL:          https://example.com
Type:         HTTP
Interval:     60s
Timeout:      10s
Regions:      us-east-1, eu-west-1, ap-southeast-1
Status:       pending

The probe will start checking within 30 seconds.
Use 'stackeye probe get {probe_id}' to view status.
```

### Understanding the Output

- **ID**: Unique identifier for your probe (use this in other commands)
- **Interval**: How often the probe runs (default: 60 seconds)
- **Timeout**: How long to wait for a response (default: 10 seconds)
- **Regions**: Where checks run from (default: all active regions)
- **Status**: Initially "pending" until the first check completes

### Probe Options

You can customize your probe with additional options:

```bash
# Custom interval and timeout
stackeye probe create \
  --name "API Health Check" \
  --url https://api.example.com/health \
  --interval 60 \
  --timeout 5

# Specific regions only
stackeye probe create \
  --name "EU Service" \
  --url https://eu.example.com \
  --regions eu-west-1,eu-central-1

# Check for specific content in response
stackeye probe create \
  --name "API Status" \
  --url https://api.example.com/status \
  --keyword-check "healthy" \
  --keyword-check-type contains

# POST request with JSON body
stackeye probe create \
  --name "Webhook Endpoint" \
  --url https://api.example.com/webhook \
  --method POST \
  --body '{"test": true}' \
  --headers '{"Content-Type": "application/json"}'
```

## Step 4: Set Up a Notification Channel

A **channel** is where StackEye sends alerts when a probe detects an issue. Let's create an email notification channel.

### Create an Email Channel

```bash
stackeye channel create \
  --name "My Email Alerts" \
  --type email \
  --email {your_email}
```

**Expected output:**

```
Channel created successfully!

ID:       {channel_id}
Name:     My Email Alerts
Type:     email
Enabled:  true

Email:    {your_email}
```

### Other Channel Types

StackEye supports multiple notification channels:

```bash
# Slack
stackeye channel create \
  --name "Slack Alerts" \
  --type slack \
  --webhook-url {your_slack_webhook_url}

# Discord
stackeye channel create \
  --name "Discord Alerts" \
  --type discord \
  --webhook-url {your_discord_webhook_url}

# Microsoft Teams
stackeye channel create \
  --name "Teams Alerts" \
  --type teams \
  --webhook-url {your_teams_webhook_url}

# PagerDuty
stackeye channel create \
  --name "On-Call Alerts" \
  --type pagerduty \
  --routing-key {your_pagerduty_routing_key} \
  --severity critical

# Custom Webhook
stackeye channel create \
  --name "Custom Integration" \
  --type webhook \
  --url {your_webhook_url} \
  --method POST
```

### Interactive Wizard

For a guided experience, use the channel wizard:

```bash
stackeye channel wizard
```

This walks you through channel creation step-by-step with prompts for each option.

## Step 5: Link the Channel to Your Probe

Now connect the notification channel to your probe so you receive alerts.

### Get Your IDs

First, find your probe and channel IDs:

```bash
# List probes
stackeye probe list

# List channels
stackeye channel list
```

### Link the Channel

```bash
stackeye probe link-channel "{probe_name_or_id}" {channel_id}
```

**Example:**

```bash
stackeye probe link-channel "My Website" {channel_id}
```

**Expected output:**

```
Successfully linked channel "My Email Alerts" to probe "My Website"

Probe: My Website
  ID:       {probe_id}
  Channels: My Email Alerts

Alerts for this probe will now be sent to the linked channel.
```

## Step 6: Monitor Your Probe

### View Probe Status

```bash
stackeye probe get "My Website"
```

**Expected output:**

```
Probe Details
─────────────────────────────────────────────

Name:           My Website
ID:             {probe_id}
URL:            https://example.com
Type:           HTTP
Status:         up
Interval:       30s
Timeout:        10s

Last Check:     {timestamp}
Response Time:  {latency_ms}ms
Status Code:    200

Linked Channels:
  - My Email Alerts ({channel_id})

Uptime (24h):   100.00%
Uptime (7d):    100.00%
Uptime (30d):   100.00%
```

### View Check History

```bash
stackeye probe history "My Website" --limit 10
```

**Expected output:**

```
TIMESTAMP                  STATUS  RESPONSE_TIME  STATUS_CODE  REGION
{timestamp}                up      {latency}ms    200          us-east-1
{timestamp}                up      {latency}ms    200          eu-west-1
{timestamp}                up      {latency}ms    200          ap-southeast-1
...
```

### View Statistics

```bash
stackeye probe stats "My Website" --period 7d
```

**Expected output:**

```
Statistics for "My Website" (Last 7 days)
─────────────────────────────────────────────

Uptime:            99.95%
Total Checks:      {count}
Successful:        {count}
Failed:            {count}

Response Times:
  Average:         {latency}ms
  P95:             {latency}ms
  P99:             {latency}ms
  Min:             {latency}ms
  Max:             {latency}ms
```

### Dashboard Overview

For a quick overview of all your monitoring:

```bash
stackeye dashboard
```

## Step 7: Handle Alerts

When a probe detects an issue (e.g., your site returns a 500 error), StackEye creates an alert and sends notifications to linked channels.

### View Active Alerts

```bash
stackeye alert list --status active
```

**Expected output:**

```
ID        PROBE          STATUS    SEVERITY   TRIGGERED           DURATION
{id}      My Website     active    critical   {timestamp}         {duration}
```

### Acknowledge an Alert

Acknowledging an alert indicates you're aware of the issue and working on it:

```bash
stackeye alert ack {alert_id}
```

### Resolve an Alert

Once the issue is fixed, resolve the alert:

```bash
stackeye alert resolve {alert_id}
```

**Note:** Alerts are also automatically resolved when the probe detects the endpoint is healthy again.

## Common Workflows

### CI/CD Integration

Monitor your deployments automatically:

```bash
# In your CI/CD pipeline (GitHub Actions, GitLab CI, etc.)
export STACKEYE_API_KEY=${{ secrets.STACKEYE_API_KEY }}

# Setup without prompts
stackeye setup --no-input

# Create probe for this deployment
stackeye probe create \
  --name "Production API - ${GITHUB_SHA:0:7}" \
  --url https://api.example.com/health
```

See the [README](../README.md#cicd-and-automation) for complete CI/CD examples.

### Multiple Probes

Create probes for different parts of your infrastructure:

```bash
# Website
stackeye probe create --name "Website" --url https://example.com

# API
stackeye probe create --name "API" --url https://api.example.com/health

# Database (TCP check)
stackeye probe create --name "Database" --url db.example.com:5432 --check-type tcp

# DNS resolution
stackeye probe create --name "DNS" --url example.com --check-type dns_resolve
```

### Bulk Operations with JSON Output

Use JSON output for scripting:

```bash
# List all probes as JSON
stackeye probe list -o json | jq '.[] | .id'

# Get probe details as JSON
stackeye probe get "My Website" -o json
```

### Multiple Organizations

If you have access to multiple organizations:

```bash
# List available organizations
stackeye org list

# Switch organizations
stackeye org switch {org_id}

# Or use contexts for quick switching
stackeye context list
stackeye context use {context_name}
```

## Troubleshooting

### "Authentication required" error

Your credentials may have expired. Re-authenticate:

```bash
stackeye login
```

### "Probe not found" error

Check the probe name or ID:

```bash
stackeye probe list
```

### Connection issues

Check your network and verify the API is reachable:

```bash
stackeye --debug whoami
```

### View detailed error information

Use the `--debug` flag for verbose output:

```bash
stackeye --debug probe create --name "Test" --url https://example.com
```

## Next Steps

Now that you have monitoring set up, explore these advanced features:

- **Status Pages**: Create public status pages for your services
  ```bash
  stackeye status-page list
  stackeye status-page create --name "Status" --slug my-status
  ```

- **Maintenance Windows**: Schedule maintenance to suppress alerts
  ```bash
  stackeye maintenance list
  stackeye maintenance create --probe-id {id} --start "{start_datetime}" --duration 2h
  ```

- **Team Management**: Invite team members to your organization
  ```bash
  stackeye team list
  stackeye team invite --email {email} --role member
  ```

- **Labels**: Organize probes with labels for filtering
  ```bash
  stackeye probe label "My Website" env=production tier=web
  stackeye probe list --labels "env=production"
  ```

## Getting Help

- **Command Help**: `stackeye <command> --help`
- **Documentation**: [docs.stackeye.io](https://docs.stackeye.io)
- **Support**: [GitHub Issues](https://github.com/StackEye-IO/stackeye-cli/issues)

## Recording Your Terminal Sessions

To create terminal recordings for documentation or demos, we recommend [asciinema](https://asciinema.org/):

```bash
# Install asciinema
brew install asciinema  # macOS
apt install asciinema   # Ubuntu/Debian

# Record a session
asciinema rec stackeye-demo.cast

# Play it back
asciinema play stackeye-demo.cast

# Upload to asciinema.org
asciinema upload stackeye-demo.cast
```

You can embed recordings in documentation using the asciinema player.
