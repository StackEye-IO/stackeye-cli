#!/bin/bash
# Post-Deployment Monitoring Script
# ==================================
#
# Use this script in your deployment pipeline to:
# 1. Create or update monitoring for the deployed service
# 2. Run immediate health checks
# 3. Verify deployment health before marking as successful
#
# Usage:
#   ./monitor-deployment.sh
#
# Environment Variables (required):
#   STACKEYE_API_KEY - Your StackEye API key
#   APP_NAME         - Application name
#   APP_URL          - Health check URL
#
# Environment Variables (optional):
#   INTERVAL         - Check interval in seconds (default: 60)
#   TIMEOUT          - Request timeout in seconds (default: 10)
#   WAIT_TIME        - Seconds to wait for checks (default: 120)
#   REQUIRED_CHECKS  - Number of successful checks required (default: 2)

set -euo pipefail

# Configuration
APP_NAME="${APP_NAME:?APP_NAME is required}"
APP_URL="${APP_URL:?APP_URL is required}"
INTERVAL="${INTERVAL:-60}"
TIMEOUT="${TIMEOUT:-10}"
WAIT_TIME="${WAIT_TIME:-120}"
REQUIRED_CHECKS="${REQUIRED_CHECKS:-2}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo "═══════════════════════════════════════════════════════════"
echo "  Post-Deployment Monitoring: ${APP_NAME}"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Verify authentication
log_info "Verifying StackEye authentication..."
if ! stackeye whoami >/dev/null 2>&1; then
    log_error "Not authenticated. Run 'stackeye login' or set STACKEYE_API_KEY"
    exit 1
fi
log_success "Authentication verified"

# Create or find probe
PROBE_NAME="${APP_NAME} - Health"
log_info "Setting up probe: ${PROBE_NAME}"

PROBE_ID=$(stackeye probe list -o json 2>/dev/null | jq -r ".[] | select(.name == \"${PROBE_NAME}\") | .id" | head -1 || echo "")

if [ -z "$PROBE_ID" ]; then
    log_info "Creating new probe..."
    PROBE_OUTPUT=$(stackeye probe create \
        --name "${PROBE_NAME}" \
        --url "${APP_URL}" \
        --interval "${INTERVAL}" \
        --timeout "${TIMEOUT}" \
        -o json 2>/dev/null || echo "{}")

    PROBE_ID=$(echo "$PROBE_OUTPUT" | jq -r '.id // empty')

    if [ -n "$PROBE_ID" ]; then
        log_success "Probe created: ${PROBE_ID}"
    else
        log_error "Failed to create probe"
        exit 1
    fi
else
    log_info "Using existing probe: ${PROBE_ID}"

    # Update probe URL in case it changed
    log_info "Updating probe URL..."
    stackeye probe update "${PROBE_ID}" --url "${APP_URL}" >/dev/null 2>&1 || true
fi

# Run immediate health check
log_info "Running immediate health check..."
if stackeye probe test "${PROBE_ID}" 2>/dev/null; then
    log_success "Immediate check passed"
else
    log_warn "Immediate check failed - this may be normal during deployment"
fi

# Wait for scheduled checks
log_info "Waiting ${WAIT_TIME}s for scheduled checks to complete..."
ELAPSED=0
SUCCESS_COUNT=0
CHECK_INTERVAL=30

while [ $ELAPSED -lt $WAIT_TIME ]; do
    sleep $CHECK_INTERVAL
    ELAPSED=$((ELAPSED + CHECK_INTERVAL))

    # Check probe status
    STATUS=$(stackeye probe get "${PROBE_ID}" -o json 2>/dev/null | jq -r '.status // "unknown"')

    if [ "$STATUS" = "up" ]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        log_success "Check ${SUCCESS_COUNT}/${REQUIRED_CHECKS} passed (status: up)"

        if [ $SUCCESS_COUNT -ge $REQUIRED_CHECKS ]; then
            break
        fi
    else
        log_warn "Check failed (status: ${STATUS})"
        SUCCESS_COUNT=0  # Reset on failure
    fi

    log_info "Elapsed: ${ELAPSED}s / ${WAIT_TIME}s"
done

# Final status check
echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  Deployment Verification Results"
echo "═══════════════════════════════════════════════════════════"
echo ""

FINAL_STATUS=$(stackeye probe get "${PROBE_ID}" -o json 2>/dev/null | jq -r '.status // "unknown"')
PROBE_DETAILS=$(stackeye probe get "${PROBE_ID}" 2>/dev/null || echo "Unable to fetch details")

echo "$PROBE_DETAILS"
echo ""

if [ "$FINAL_STATUS" = "up" ] && [ $SUCCESS_COUNT -ge $REQUIRED_CHECKS ]; then
    log_success "Deployment verification PASSED"
    echo ""
    echo "The service is responding correctly to health checks."
    echo "Probe ID: ${PROBE_ID}"
    exit 0
else
    log_error "Deployment verification FAILED"
    echo ""
    echo "The service did not pass health checks."
    echo "Status: ${FINAL_STATUS}"
    echo "Successful checks: ${SUCCESS_COUNT}/${REQUIRED_CHECKS}"
    echo ""
    echo "Check the probe history for details:"
    echo "  stackeye probe history ${PROBE_ID}"
    echo ""
    echo "View active alerts:"
    echo "  stackeye alert list --status active"
    exit 1
fi
