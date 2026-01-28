// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// statusPageGetStatusTimeout is the maximum time to wait for the API response.
const statusPageGetStatusTimeout = 30 * time.Second

// NewStatusPageGetStatusCmd creates and returns the status-page get-status subcommand.
func NewStatusPageGetStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-status <id>",
		Short: "Get aggregated status of a status page",
		Long: `Get the current aggregated status of a status page and its probes.

Displays the overall health status of the status page along with the status
of each probe assigned to it. This provides a quick overview of which components
are operational, degraded, or experiencing issues.

Output Fields:
  OVERALL STATUS     Aggregated status (Operational, Degraded, Outage)
  LAST UPDATED       When the status was last calculated

Probe Status Fields:
  NAME               Display name of the probe
  STATUS             Current status (up, down, degraded, pending)
  UPTIME             Uptime percentage (if enabled for display)
  RESPONSE TIME      Average response time in ms (if enabled)

Wide Mode Fields (--output wide):
  PROBE ID           UUID of the probe

Status Values:
  Operational        All probes are UP
  Degraded           Some probes are DOWN or degraded
  Outage             All probes are DOWN

Examples:
  # Get current status of a status page
  stackeye status-page get-status 123

  # Output as JSON for scripting
  stackeye status-page get-status 123 -o json

  # Output as YAML
  stackeye status-page get-status 123 -o yaml

  # Wide output with probe IDs
  stackeye status-page get-status 123 -o wide`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageGetStatus(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runStatusPageGetStatus executes the status-page get-status command logic.
func runStatusPageGetStatus(ctx context.Context, idArg string) error {
	// Parse and validate status page ID (uint)
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}
	if id == 0 {
		return fmt.Errorf("invalid status page ID: must be greater than 0")
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get aggregated status with timeout
	reqCtx, cancel := context.WithTimeout(ctx, statusPageGetStatusTimeout)
	defer cancel()

	status, err := client.GetAggregatedStatus(reqCtx, apiClient, uint(id))
	if err != nil {
		return fmt.Errorf("failed to get status page status: %w", err)
	}

	// Defensive check for nil status
	if status == nil {
		return fmt.Errorf("status page %d not found or has no status", id)
	}

	// Print the aggregated status using the configured output format
	return output.PrintAggregatedStatus(*status)
}
