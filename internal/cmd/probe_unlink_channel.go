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

// probeUnlinkChannelTimeout is the maximum time to wait for the API response.
const probeUnlinkChannelTimeout = 30 * time.Second

// NewProbeUnlinkChannelCmd creates and returns the probe unlink-channel subcommand.
func NewProbeUnlinkChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "unlink-channel <probe-id> <channel-id>",
		Short:             "Unlink a notification channel from a probe",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Unlink a notification channel from a probe to stop receiving alert notifications.

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead. The channel must still
be specified by UUID.

After unlinking, the probe will no longer send alerts to this channel.
Other linked channels will continue to receive notifications.

Examples:
  # Unlink a channel from a probe by name
  stackeye probe unlink-channel "Production API" \
    660e8400-e29b-41d4-a716-446655440001

  # Unlink a channel from a probe by UUID
  stackeye probe unlink-channel 550e8400-e29b-41d4-a716-446655440000 \
    660e8400-e29b-41d4-a716-446655440001

Use 'stackeye probe get <probe-id>' to view currently linked channels.
Use 'stackeye channel list' to see available notification channels.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeUnlinkChannel(cmd, args[0], args[1])
		},
	}

	return cmd
}

// runProbeUnlinkChannel executes the probe unlink-channel command logic.
func runProbeUnlinkChannel(cmd *cobra.Command, probeIDArg, channelIDArg string) error {
	// Parse and validate channel UUID (channels are always referenced by UUID)
	channelID, err := uuid.Parse(channelIDArg)
	if err != nil {
		return fmt.Errorf("invalid channel ID %q: must be a valid UUID", channelIDArg)
	}

	// Dry-run check: print what would happen and exit without making API calls
	if GetDryRun() {
		dryrun.PrintAction("unlink channel from", "probe",
			"Probe", probeIDArg,
			"Channel", channelIDArg,
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Resolve probe ID (accepts UUID or name)
	probeID, err := ResolveProbeID(cmd.Context(), apiClient, probeIDArg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), probeUnlinkChannelTimeout)
	defer cancel()

	// Get the probe to retrieve current channels
	probe, err := client.GetProbe(ctx, apiClient, probeID, "")
	if err != nil {
		return fmt.Errorf("failed to get probe: %w", err)
	}

	// Validate channel exists
	channel, err := client.GetChannel(ctx, apiClient, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Check if channel is actually linked
	channelIndex := -1
	for i, existingID := range probe.AlertChannelIDs {
		if existingID == channelID {
			channelIndex = i
			break
		}
	}

	if channelIndex == -1 {
		return fmt.Errorf("channel %q is not linked to probe %q", channel.Name, probe.Name)
	}

	// Build update request with channel removed
	newChannelIDs := make([]uuid.UUID, 0, len(probe.AlertChannelIDs)-1)
	for i, id := range probe.AlertChannelIDs {
		if i != channelIndex {
			newChannelIDs = append(newChannelIDs, id)
		}
	}
	req := &client.UpdateProbeRequest{
		AlertChannelIDs: newChannelIDs,
	}

	// Update the probe with the new channel list
	updatedProbe, err := client.UpdateProbe(ctx, apiClient, probeID, req)
	if err != nil {
		return fmt.Errorf("failed to unlink channel: %w", err)
	}

	// Print success message and updated probe
	fmt.Printf("Successfully unlinked channel %q from probe %q\n\n", channel.Name, probe.Name)
	return output.Print(updatedProbe)
}
