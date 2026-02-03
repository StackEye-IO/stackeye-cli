#!/bin/bash
# StackEye Bootstrap Script
# =========================
#
# One-command setup for StackEye monitoring.
# Installs CLI, authenticates, and creates initial monitoring configuration.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/StackEye-IO/stackeye-cli/main/examples/scripts/bootstrap.sh | bash
#
# Or locally:
#   chmod +x bootstrap.sh
#   ./bootstrap.sh
#
# Environment Variables:
#   STACKEYE_API_KEY  - API key for non-interactive setup
#   APP_NAME          - Application name for probes (default: hostname)
#   APP_URL           - URL to monitor (required)
#   SLACK_WEBHOOK_URL - Slack webhook for notifications (optional)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Banner
echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║           StackEye Monitoring Bootstrap Script            ║"
echo "║                 https://stackeye.io                       ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Check dependencies
log_info "Checking dependencies..."
command -v curl >/dev/null 2>&1 || { log_error "curl is required but not installed."; exit 1; }
command -v jq >/dev/null 2>&1 || { log_warn "jq not found - some features may be limited"; }

# Install StackEye CLI
log_info "Installing StackEye CLI..."
if command -v stackeye >/dev/null 2>&1; then
    log_success "StackEye CLI already installed"
    stackeye version
else
    curl -fsSL https://get.stackeye.io/cli | bash
    export PATH="$PATH:$HOME/.local/bin"
    log_success "StackEye CLI installed"
fi

# Authenticate
log_info "Configuring authentication..."
if [ -n "${STACKEYE_API_KEY:-}" ]; then
    stackeye setup --no-input
    log_success "Configured with API key"
else
    log_info "No API key provided. Starting interactive login..."
    stackeye login
fi

# Verify authentication
log_info "Verifying authentication..."
stackeye whoami
log_success "Authentication verified"

# Get configuration
APP_NAME="${APP_NAME:-$(hostname)}"
APP_URL="${APP_URL:-}"

# Prompt for URL if not provided
if [ -z "$APP_URL" ]; then
    echo ""
    read -p "Enter the URL to monitor (e.g., https://api.example.com/health): " APP_URL
fi

if [ -z "$APP_URL" ]; then
    log_error "APP_URL is required"
    exit 1
fi

# Create probe
log_info "Creating monitoring probe for ${APP_NAME}..."
PROBE_OUTPUT=$(stackeye probe create \
    --name "${APP_NAME} - Health Check" \
    --url "${APP_URL}" \
    --interval 60 \
    --timeout 10 \
    -o json 2>/dev/null || echo "{}")

if command -v jq >/dev/null 2>&1 && [ -n "$PROBE_OUTPUT" ] && [ "$PROBE_OUTPUT" != "{}" ]; then
    PROBE_ID=$(echo "$PROBE_OUTPUT" | jq -r '.id // empty')
    if [ -n "$PROBE_ID" ]; then
        log_success "Probe created with ID: $PROBE_ID"
    fi
else
    log_warn "Probe may already exist or creation had issues"
    # Try to find existing probe
    PROBE_ID=$(stackeye probe list -o json | jq -r ".[] | select(.name | contains(\"${APP_NAME}\")) | .id" | head -1 2>/dev/null || echo "")
    if [ -n "$PROBE_ID" ]; then
        log_info "Found existing probe: $PROBE_ID"
    fi
fi

# Create Slack channel if webhook provided
if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
    log_info "Creating Slack notification channel..."

    cat > /tmp/stackeye-slack.yaml << EOF
name: "${APP_NAME} - Slack Alerts"
type: slack
enabled: true
config:
  webhook_url: "${SLACK_WEBHOOK_URL}"
EOF

    if stackeye channel create --from-file /tmp/stackeye-slack.yaml 2>/dev/null; then
        log_success "Slack channel created"

        # Link to probe
        if [ -n "$PROBE_ID" ]; then
            CHANNEL_ID=$(stackeye channel list -o json | jq -r ".[] | select(.name | contains(\"${APP_NAME}\")) | .id" | head -1 2>/dev/null || echo "")
            if [ -n "$CHANNEL_ID" ]; then
                stackeye probe link-channel "$PROBE_ID" "$CHANNEL_ID" 2>/dev/null && \
                    log_success "Slack channel linked to probe"
            fi
        fi
    else
        log_warn "Slack channel may already exist"
    fi

    rm -f /tmp/stackeye-slack.yaml
fi

# Run initial check
log_info "Running initial health check..."
if [ -n "$PROBE_ID" ]; then
    stackeye probe test "$PROBE_ID" || log_warn "Initial check may have failed - this is normal if service is still starting"
fi

# Summary
echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                    Setup Complete!                        ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
log_success "StackEye monitoring is now active"
echo ""
echo "Next steps:"
echo "  • View your probes:    stackeye probe list"
echo "  • View active alerts:  stackeye alert list --status active"
echo "  • Check probe status:  stackeye probe get \"${PROBE_ID:-YOUR_PROBE_ID}\""
echo "  • View dashboard:      stackeye dashboard"
echo ""
echo "Documentation: https://docs.stackeye.io"
echo "Support:       https://github.com/StackEye-IO/stackeye-cli/issues"
echo ""
