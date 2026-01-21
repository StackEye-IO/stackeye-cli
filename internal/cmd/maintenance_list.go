// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// maintenanceListTimeout is the maximum time to wait for the API response.
const maintenanceListTimeout = 30 * time.Second

// maintenanceListFlags holds the flag values for the maintenance list command.
type maintenanceListFlags struct {
	includeExpired bool
	page           int
	limit          int
}

// NewMaintenanceListCmd creates and returns the maintenance list subcommand.
func NewMaintenanceListCmd() *cobra.Command {
	flags := &maintenanceListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all maintenance windows",
		Long: `List all scheduled maintenance windows in your organization.

Displays maintenance window status, scope, target, duration, expiration time,
name, and reason. By default, only active (non-expired) windows are shown.

Maintenance windows are named periods during which alert notifications are
suppressed. They provide better tracking than ad-hoc mutes for planned
downtime such as deployments, upgrades, or infrastructure changes.

Scope Types:
  Organization  Maintenance applies to entire organization
  Probe         Maintenance for a specific probe

Status Values:
  ACTIVE        Maintenance window is currently in effect
  EXPIRED       Maintenance window has ended

Examples:
  # List all active maintenance windows
  stackeye maintenance list

  # Include expired maintenance windows
  stackeye maintenance list --include-expired

  # Output as JSON for scripting
  stackeye maintenance list -o json

  # Wide output with additional columns (ID, timestamps)
  stackeye maintenance list -o wide

  # Paginate through results
  stackeye maintenance list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMaintenanceList(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().BoolVar(&flags.includeExpired, "include-expired", false, "include expired maintenance windows")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")

	return cmd
}

// runMaintenanceList executes the maintenance list command logic.
func runMaintenanceList(ctx context.Context, flags *maintenanceListFlags) error {
	// Validate all flags before making any API calls
	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options - always filter for maintenance windows only
	// SDK uses offset-based pagination, convert page to offset
	offset := (flags.page - 1) * flags.limit
	opts := &client.ListMutesOptions{
		Limit:           flags.limit,
		Offset:          offset,
		IncludeExpired:  flags.includeExpired,
		MaintenanceOnly: true, // Always filter for maintenance windows
	}

	// Call SDK to list maintenance windows with timeout
	reqCtx, cancel := context.WithTimeout(ctx, maintenanceListTimeout)
	defer cancel()

	result, err := client.ListMutes(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list maintenance windows: %w", err)
	}

	// Handle empty results
	if len(result.Data) == 0 {
		msg := "No active maintenance windows found"
		if flags.includeExpired {
			msg = "No maintenance windows found"
		}
		return output.PrintEmpty(msg)
	}

	// Print the maintenance windows using the mute output format
	return output.PrintMutes(result.Data)
}
