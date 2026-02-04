// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// statusPageAddProbeTimeout is the maximum time to wait for the add-probe API response.
const statusPageAddProbeTimeout = 30 * time.Second

// statusPageAddProbeFlags holds the flag values for the status-page add-probe command.
type statusPageAddProbeFlags struct {
	probeID          string
	displayName      string
	showResponseTime bool
}

// NewStatusPageAddProbeCmd creates and returns the status-page add-probe subcommand.
func NewStatusPageAddProbeCmd() *cobra.Command {
	flags := &statusPageAddProbeFlags{}

	cmd := &cobra.Command{
		Use:   "add-probe <status-page-id>",
		Short: "Add a probe to a status page",
		Long: `Add a probe to a status page for public display.

This command associates an existing probe with a status page, making the probe's
status visible on the public status page. Each probe can have a custom display
name for user-friendly presentation.

The probe must already exist and belong to the same organization as the status page.

Flags:
  --probe-id        Required. The UUID of the probe to add.
  --display-name    Optional. Custom name shown on the status page.
                    If not provided, the probe's original name is used.
  --show-response-time  Optional. Show response time metrics on the status page.
                        Defaults to false.

Examples:
  # Add a probe to status page 123
  stackeye status-page add-probe 123 --probe-id {probe_uuid}

  # Add a probe with a custom display name
  stackeye status-page add-probe 123 --probe-id {probe_uuid} --display-name "API Server"

  # Add a probe and show response time
  stackeye status-page add-probe 123 --probe-id {probe_uuid} --show-response-time`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageAddProbe(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().StringVar(&flags.probeID, "probe-id", "", "probe UUID to add (required)")
	cmd.Flags().StringVar(&flags.displayName, "display-name", "", "custom display name for the probe on the status page")
	cmd.Flags().BoolVar(&flags.showResponseTime, "show-response-time", false, "show response time metrics on the status page")

	_ = cmd.MarkFlagRequired("probe-id")

	return cmd
}

// runStatusPageAddProbe executes the status-page add-probe command logic.
func runStatusPageAddProbe(ctx context.Context, idArg string, flags *statusPageAddProbeFlags) error {
	// Parse and validate status page ID
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}

	// Validate probe ID is provided and is a valid UUID
	if flags.probeID == "" {
		return fmt.Errorf("--probe-id is required")
	}

	parsedUUID, err := uuid.Parse(flags.probeID)
	if err != nil {
		return fmt.Errorf("invalid probe ID %q: must be a valid UUID", flags.probeID)
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		details := []string{
			"Status Page ID", idArg,
			"Probe ID", parsedUUID.String(),
		}
		if flags.displayName != "" {
			details = append(details, "Display Name", flags.displayName)
		}
		dryrun.PrintAction("add probe to", "status page", details...)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build the request with normalized UUID
	req := &client.AddProbeToStatusPageRequest{
		ProbeID:          parsedUUID.String(),
		ShowResponseTime: flags.showResponseTime,
	}

	// Set display name if provided
	if flags.displayName != "" {
		req.DisplayName = &flags.displayName
	}

	// Add the probe to the status page
	addCtx, cancel := context.WithTimeout(ctx, statusPageAddProbeTimeout)
	probe, err := client.AddProbeToStatusPage(addCtx, apiClient, uint(id), req)
	cancel()

	if err != nil {
		return fmt.Errorf("failed to add probe to status page: %w", err)
	}

	// Display success message
	displayName := flags.probeID
	if probe.DisplayName != nil && *probe.DisplayName != "" {
		displayName = *probe.DisplayName
	}
	fmt.Printf("Added probe %s to status page %d\n", displayName, id)

	return nil
}
