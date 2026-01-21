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

// muteListTimeout is the maximum time to wait for the API response.
const muteListTimeout = 30 * time.Second

// muteListFlags holds the flag values for the mute list command.
type muteListFlags struct {
	includeExpired  bool
	maintenanceOnly bool
	page            int
	limit           int
}

// NewMuteListCmd creates and returns the mute list subcommand.
func NewMuteListCmd() *cobra.Command {
	flags := &muteListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all alert mute periods",
		Long: `List all alert mute periods in your organization.

Displays mute status, scope type, target, duration, expiration time, and reason.
By default, only active (non-expired) mutes are shown.

Scope Types:
  Organization  Silence all alerts for the entire organization
  Probe         Silence alerts for a specific probe
  Channel       Silence a specific notification channel
  Alert Type    Silence alerts of a specific type

Status Values:
  ACTIVE        Mute is currently in effect
  EXPIRED       Mute has ended

Examples:
  # List all active mutes
  stackeye mute list

  # Include expired mutes in the list
  stackeye mute list --include-expired

  # List only maintenance windows
  stackeye mute list --maintenance-only

  # List maintenance windows including expired ones
  stackeye mute list --maintenance-only --include-expired

  # Output as JSON for scripting
  stackeye mute list -o json

  # Wide output with additional columns
  stackeye mute list -o wide

  # Paginate through results
  stackeye mute list --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMuteList(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().BoolVar(&flags.includeExpired, "include-expired", false, "include expired mutes in the list")
	cmd.Flags().BoolVar(&flags.maintenanceOnly, "maintenance-only", false, "only show maintenance windows")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")

	return cmd
}

// runMuteList executes the mute list command logic.
func runMuteList(ctx context.Context, flags *muteListFlags) error {
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

	// Build list options from validated flags
	// SDK uses offset-based pagination, convert page to offset
	offset := (flags.page - 1) * flags.limit
	opts := &client.ListMutesOptions{
		Limit:           flags.limit,
		Offset:          offset,
		IncludeExpired:  flags.includeExpired,
		MaintenanceOnly: flags.maintenanceOnly,
	}

	// Call SDK to list mutes with timeout
	reqCtx, cancel := context.WithTimeout(ctx, muteListTimeout)
	defer cancel()

	result, err := client.ListMutes(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list mutes: %w", err)
	}

	// Handle empty results
	if len(result.Data) == 0 {
		msg := "No mutes found"
		if !flags.includeExpired {
			msg = "No active mutes found"
		}
		if flags.maintenanceOnly {
			msg = "No maintenance windows found"
		}
		return output.PrintEmpty(msg)
	}

	// Print the mutes using the configured output format
	return output.PrintMutes(result.Data)
}
