// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// statusPageReorderProbesTimeout is the maximum time to wait for the reorder-probes API response.
const statusPageReorderProbesTimeout = 30 * time.Second

// statusPageReorderProbesFlags holds the flag values for the status-page reorder-probes command.
type statusPageReorderProbesFlags struct {
	probeIDs string
}

// NewStatusPageReorderProbesCmd creates and returns the status-page reorder-probes subcommand.
func NewStatusPageReorderProbesCmd() *cobra.Command {
	flags := &statusPageReorderProbesFlags{}

	cmd := &cobra.Command{
		Use:   "reorder-probes <status-page-id>",
		Short: "Reorder probes on a status page",
		Long: `Reorder the display order of probes on a status page.

This command updates the display order of probes on a status page. The order
is determined by the position in the comma-separated list: the first probe ID
will have order 0, the second will have order 1, and so on.

All probes currently on the status page should be included in the list. Probes
not included will retain their current order values but may appear after the
reordered probes.

Flags:
  --probe-ids    Required. Comma-separated list of probe UUIDs in desired order.
                 No spaces between UUIDs.

Examples:
  # Reorder probes on status page 123
  stackeye status-page reorder-probes 123 --probe-ids {uuid1},{uuid2},{uuid3}

  # Put the API probe first, then Database, then Website
  stackeye status-page reorder-probes 123 \
    --probe-ids 550e8400-e29b-41d4-a716-446655440001,550e8400-e29b-41d4-a716-446655440002,550e8400-e29b-41d4-a716-446655440003`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageReorderProbes(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().StringVar(&flags.probeIDs, "probe-ids", "", "comma-separated list of probe UUIDs in desired order (required)")

	_ = cmd.MarkFlagRequired("probe-ids")

	return cmd
}

// runStatusPageReorderProbes executes the status-page reorder-probes command logic.
func runStatusPageReorderProbes(ctx context.Context, idArg string, flags *statusPageReorderProbesFlags) error {
	// Parse and validate status page ID
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}

	// Validate probe IDs are provided
	if flags.probeIDs == "" {
		return fmt.Errorf("--probe-ids is required")
	}

	// Parse and validate the comma-separated probe UUIDs
	probeIDStrings := strings.Split(flags.probeIDs, ",")
	orders := make([]client.ProbeOrderItem, 0, len(probeIDStrings))
	orderIndex := 0

	for i, probeIDStr := range probeIDStrings {
		probeIDStr = strings.TrimSpace(probeIDStr)
		if probeIDStr == "" {
			continue // Skip empty entries from trailing commas
		}

		if _, err := uuid.Parse(probeIDStr); err != nil {
			return fmt.Errorf("invalid probe ID %q at position %d: must be a valid UUID", probeIDStr, i+1)
		}

		orders = append(orders, client.ProbeOrderItem{
			ProbeID: probeIDStr,
			Order:   orderIndex,
		})
		orderIndex++
	}

	if len(orders) == 0 {
		return fmt.Errorf("--probe-ids must contain at least one valid probe UUID")
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		ids := make([]string, len(orders))
		for i, o := range orders {
			ids[i] = o.ProbeID
		}
		dryrun.PrintAction("reorder probes on", "status page",
			"Status Page ID", idArg,
			"Probe Count", fmt.Sprintf("%d", len(orders)),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build the request
	req := &client.ReorderProbesRequest{
		Orders: orders,
	}

	// Reorder the probes
	reorderCtx, cancel := context.WithTimeout(ctx, statusPageReorderProbesTimeout)
	err = client.ReorderProbes(reorderCtx, apiClient, uint(id), req)
	cancel()

	if err != nil {
		return fmt.Errorf("failed to reorder probes: %w", err)
	}

	// Display success message
	fmt.Printf("Reordered %d probes on status page %d\n", len(orders), id)

	return nil
}
