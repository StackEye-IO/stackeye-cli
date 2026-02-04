// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// maintenanceCreateTimeout is the maximum time to wait for the API response.
const maintenanceCreateTimeout = 30 * time.Second

// maintenanceCreateFlags holds the flag values for the maintenance create command.
type maintenanceCreateFlags struct {
	// Required
	name     string
	duration int

	// Scope selection
	probeID          string
	organizationWide bool

	// Optional
	reason   string
	startsAt string
}

// NewMaintenanceCreateCmd creates and returns the maintenance create subcommand.
func NewMaintenanceCreateCmd() *cobra.Command {
	flags := &maintenanceCreateFlags{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Schedule a new maintenance window",
		Long: `Schedule a new maintenance window to suppress alerts during planned downtime.

Maintenance windows are named periods during which alert notifications are
suppressed. Unlike ad-hoc mutes, maintenance windows are designed for planned
downtime and provide:
  - Named windows for audit trails and team coordination
  - Scheduled start times for future maintenance
  - Clear scope definition (specific probe or organization-wide)

Required Flags:
  --name      A descriptive name for the maintenance window
  --duration  Duration in minutes (how long the window should last)

Scope Flags (choose one):
  --probe-id          Maintenance for a specific probe (requires UUID)
  --organization-wide Maintenance applies to entire organization

Optional Flags:
  --reason     Additional details about the maintenance
  --starts-at  When the window starts (RFC3339 format, default: now)

Examples:
  # Schedule a 2-hour maintenance window for a specific probe
  stackeye maintenance create --name "Server Upgrade" \
    --probe-id <uuid> --duration 120

  # Schedule organization-wide maintenance starting immediately
  stackeye maintenance create --name "Network Migration" \
    --organization-wide --duration 60

  # Schedule future maintenance with reason
  stackeye maintenance create --name "Database Maintenance" \
    --probe-id <uuid> --duration 180 \
    --starts-at 2024-01-15T02:00:00Z \
    --reason "Scheduled PostgreSQL upgrade to version 16"

  # Quick 30-minute maintenance window
  stackeye maintenance create --name "Hotfix Deploy" \
    --probe-id <uuid> --duration 30`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMaintenanceCreate(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().StringVar(&flags.name, "name", "", "maintenance window name (required)")
	cmd.Flags().IntVar(&flags.duration, "duration", 0, "duration in minutes (required)")

	// Scope flags
	cmd.Flags().StringVar(&flags.probeID, "probe-id", "", "probe UUID for probe-specific maintenance")
	cmd.Flags().BoolVar(&flags.organizationWide, "organization-wide", false, "apply maintenance to entire organization")

	// Optional flags
	cmd.Flags().StringVar(&flags.reason, "reason", "", "reason for the maintenance window")
	cmd.Flags().StringVar(&flags.startsAt, "starts-at", "", "when the window starts (RFC3339 format, default: now)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("duration")

	return cmd
}

// runMaintenanceCreate executes the maintenance create command logic.
func runMaintenanceCreate(ctx context.Context, flags *maintenanceCreateFlags) error {
	// Build the request from flags
	req, err := buildMaintenanceRequest(flags)
	if err != nil {
		return err
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		scope := "organization"
		if flags.probeID != "" {
			scope = "probe"
		}
		dryrun.PrintAction("create", "maintenance window",
			"Name", flags.name,
			"Scope", scope,
			"Duration", fmt.Sprintf("%d minutes", flags.duration),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to create maintenance window (via CreateMute with IsMaintenanceWindow=true)
	reqCtx, cancel := context.WithTimeout(ctx, maintenanceCreateTimeout)
	defer cancel()

	mute, err := client.CreateMute(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create maintenance window: %w", err)
	}

	// Print the created maintenance window using the configured output format
	return output.PrintMute(*mute)
}

// buildMaintenanceRequest constructs the API request from command flags.
// Maintenance windows are implemented as mutes with IsMaintenanceWindow=true.
func buildMaintenanceRequest(flags *maintenanceCreateFlags) (*client.CreateMuteRequest, error) {
	// Validate required fields
	if flags.name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	if flags.duration <= 0 {
		return nil, fmt.Errorf("--duration must be a positive number of minutes")
	}

	// Validate scope: must specify exactly one of --probe-id or --organization-wide
	hasProbeID := flags.probeID != ""
	hasOrgWide := flags.organizationWide

	if !hasProbeID && !hasOrgWide {
		return nil, fmt.Errorf("must specify either --probe-id or --organization-wide")
	}
	if hasProbeID && hasOrgWide {
		return nil, fmt.Errorf("cannot specify both --probe-id and --organization-wide")
	}

	// Build the request - maintenance windows are mutes with special flags
	req := &client.CreateMuteRequest{
		DurationMinutes:     flags.duration,
		IsMaintenanceWindow: true,
		MaintenanceName:     &flags.name,
	}

	// Set scope based on flags
	if hasOrgWide {
		req.ScopeType = client.MuteScopeOrganization
	} else {
		req.ScopeType = client.MuteScopeProbe
		probeUUID, err := uuid.Parse(flags.probeID)
		if err != nil {
			return nil, fmt.Errorf("invalid --probe-id: must be a valid UUID: %w", err)
		}
		req.ProbeID = &probeUUID
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

	return req, nil
}
