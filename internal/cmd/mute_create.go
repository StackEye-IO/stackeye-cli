// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	clierrors "github.com/StackEye-IO/stackeye-cli/internal/errors"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// muteCreateTimeout is the maximum time to wait for the API response.
const muteCreateTimeout = 30 * time.Second

// muteCreateFlags holds the flag values for the mute create command.
type muteCreateFlags struct {
	// Required
	scope    string
	duration int

	// Scope-specific
	probeID   string
	channelID string
	alertType string

	// Optional
	reason          string
	startsAt        string
	maintenance     bool
	maintenanceName string
}

// NewMuteCreateCmd creates and returns the mute create subcommand.
func NewMuteCreateCmd() *cobra.Command {
	flags := &muteCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new alert mute period",
		Long: `Create a new alert mute period to temporarily silence notifications.

Mutes can target different scopes:
  organization  Silence all alerts for the entire organization
  probe         Silence alerts for a specific probe (requires --probe-id)
  channel       Silence a specific notification channel (requires --channel-id)
  alert_type    Silence alerts of a specific type (requires --alert-type)

Alert Types (for --alert-type):
  status_down         Service is down/unreachable
  ssl_expiry          SSL certificate is expiring soon
  ssl_invalid         SSL certificate is invalid
  slow_response       Response time exceeded threshold
  domain_expiry       Domain registration is expiring soon
  dns_record_missing  Expected DNS record not found
  dns_record_mismatch DNS record value doesn't match expected
  security_headers    Missing or misconfigured security headers
  cert_transparency   Certificate transparency log anomaly detected

Required Flags:
  --scope       The scope of the mute (organization, probe, channel, alert_type)
  --duration    Duration in minutes (how long the mute should last)

Scope-Specific Flags:
  --probe-id    Probe UUID (required when scope is "probe")
  --channel-id  Channel UUID (required when scope is "channel")
  --alert-type  Alert type (required when scope is "alert_type")

Optional Flags:
  --reason           Reason for creating the mute
  --starts-at        When the mute should start (RFC3339 format, default: now)
  --maintenance      Mark as a maintenance window
  --maintenance-name Name for the maintenance window

Examples:
  # Mute all alerts organization-wide for 1 hour
  stackeye mute create --scope organization --duration 60

  # Mute alerts for a specific probe for 2 hours
  stackeye mute create --scope probe --probe-id <uuid> --duration 120

  # Mute a notification channel for 30 minutes
  stackeye mute create --scope channel --channel-id <uuid> --duration 30 \
    --reason "Testing channel configuration"

  # Mute all SSL expiry alerts for 24 hours
  stackeye mute create --scope alert_type --alert-type ssl_expiry --duration 1440

  # Create a scheduled maintenance window starting tomorrow
  stackeye mute create --scope probe --probe-id <uuid> --duration 120 \
    --starts-at 2024-01-15T02:00:00Z \
    --maintenance --maintenance-name "Server upgrade"

  # Mute with a reason for audit trail
  stackeye mute create --scope organization --duration 60 \
    --reason "Deploying new version, expecting brief downtime"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMuteCreate(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().StringVar(&flags.scope, "scope", "", "mute scope: organization, probe, channel, alert_type")
	cmd.Flags().IntVar(&flags.duration, "duration", 0, "duration in minutes")

	// Scope-specific flags
	cmd.Flags().StringVar(&flags.probeID, "probe-id", "", "probe UUID (required for scope=probe)")
	cmd.Flags().StringVar(&flags.channelID, "channel-id", "", "channel UUID (required for scope=channel)")
	cmd.Flags().StringVar(&flags.alertType, "alert-type", "", "alert type: status_down, ssl_expiry, ssl_invalid, slow_response, domain_expiry, dns_record_missing, dns_record_mismatch, security_headers, cert_transparency (required for scope=alert_type)")

	// Optional flags
	cmd.Flags().StringVar(&flags.reason, "reason", "", "reason for the mute")
	cmd.Flags().StringVar(&flags.startsAt, "starts-at", "", "when the mute starts (RFC3339 format, default: now)")
	cmd.Flags().BoolVar(&flags.maintenance, "maintenance", false, "mark as maintenance window")
	cmd.Flags().StringVar(&flags.maintenanceName, "maintenance-name", "", "name for the maintenance window")

	// Mark required flags
	_ = cmd.MarkFlagRequired("scope")
	_ = cmd.MarkFlagRequired("duration")

	return cmd
}

// runMuteCreate executes the mute create command logic.
func runMuteCreate(ctx context.Context, flags *muteCreateFlags) error {
	// Build the request from flags
	req, err := buildMuteRequestFromFlags(flags)
	if err != nil {
		return err
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		dryrun.PrintAction("create", "mute",
			"Scope", string(req.ScopeType),
			"Duration", fmt.Sprintf("%d minutes", req.DurationMinutes),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to create mute with timeout
	reqCtx, cancel := context.WithTimeout(ctx, muteCreateTimeout)
	defer cancel()

	mute, err := client.CreateMute(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create mute: %w", err)
	}

	// Print the created mute using the configured output format
	return output.PrintMute(*mute)
}

// buildMuteRequestFromFlags constructs the API request from command flags.
func buildMuteRequestFromFlags(flags *muteCreateFlags) (*client.CreateMuteRequest, error) {
	// Validate required fields
	if flags.scope == "" {
		return nil, fmt.Errorf("--scope is required")
	}
	if flags.duration <= 0 {
		return nil, fmt.Errorf("--duration must be a positive number of minutes")
	}

	// Validate and convert scope type
	scopeType := client.MuteScopeType(strings.ToLower(flags.scope))
	if err := validateMuteScopeType(scopeType); err != nil {
		return nil, err
	}

	// Build the request
	req := &client.CreateMuteRequest{
		ScopeType:           scopeType,
		DurationMinutes:     flags.duration,
		IsMaintenanceWindow: flags.maintenance,
	}

	// Set scope-specific fields
	if err := setMuteScopeFields(req, scopeType, flags); err != nil {
		return nil, err
	}

	// Set optional fields
	if flags.reason != "" {
		req.Reason = &flags.reason
	}

	if flags.startsAt != "" {
		t, err := time.Parse(time.RFC3339, flags.startsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid --starts-at format: must be RFC3339 (e.g., 2024-01-15T02:00:00Z): %w", err)
		}
		req.StartsAt = &t
	}

	if flags.maintenanceName != "" {
		req.MaintenanceName = &flags.maintenanceName
	}

	return req, nil
}

// validateMuteScopeType validates the mute scope type value.
func validateMuteScopeType(s client.MuteScopeType) error {
	valid := map[client.MuteScopeType]bool{
		client.MuteScopeOrganization: true,
		client.MuteScopeProbe:        true,
		client.MuteScopeChannel:      true,
		client.MuteScopeAlertType:    true,
	}
	if !valid[s] {
		return clierrors.InvalidValueError("--scope", string(s), clierrors.ValidMuteScopes)
	}
	return nil
}

// setMuteScopeFields sets scope-specific fields on the request and validates required fields.
func setMuteScopeFields(req *client.CreateMuteRequest, scopeType client.MuteScopeType, flags *muteCreateFlags) error {
	switch scopeType {
	case client.MuteScopeOrganization:
		// No additional fields required
		return nil

	case client.MuteScopeProbe:
		if flags.probeID == "" {
			return fmt.Errorf("--probe-id is required when scope is \"probe\"")
		}
		probeUUID, err := uuid.Parse(flags.probeID)
		if err != nil {
			return fmt.Errorf("invalid --probe-id: must be a valid UUID: %w", err)
		}
		req.ProbeID = &probeUUID
		return nil

	case client.MuteScopeChannel:
		if flags.channelID == "" {
			return fmt.Errorf("--channel-id is required when scope is \"channel\"")
		}
		channelUUID, err := uuid.Parse(flags.channelID)
		if err != nil {
			return fmt.Errorf("invalid --channel-id: must be a valid UUID: %w", err)
		}
		req.ChannelID = &channelUUID
		return nil

	case client.MuteScopeAlertType:
		if flags.alertType == "" {
			return fmt.Errorf("--alert-type is required when scope is \"alert_type\"")
		}
		alertType := client.AlertType(strings.ToLower(flags.alertType))
		if err := validateMuteAlertType(alertType); err != nil {
			return err
		}
		req.AlertType = &alertType
		return nil

	default:
		return fmt.Errorf("unsupported scope type: %s", scopeType)
	}
}

// validateMuteAlertType validates the alert type value.
func validateMuteAlertType(t client.AlertType) error {
	valid := map[client.AlertType]bool{
		client.AlertTypeStatusDown:        true,
		client.AlertTypeSSLExpiry:         true,
		client.AlertTypeSSLInvalid:        true,
		client.AlertTypeSlowResponse:      true,
		client.AlertTypeDomainExpiry:      true,
		client.AlertTypeDNSRecordMissing:  true,
		client.AlertTypeDNSRecordMismatch: true,
		client.AlertTypeSecurityHeaders:   true,
		client.AlertTypeCertTransparency:  true,
	}
	if !valid[t] {
		return clierrors.InvalidValueError("--alert-type", string(t), clierrors.ValidMuteAlertTypes)
	}
	return nil
}
