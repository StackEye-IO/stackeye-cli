// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// probeLinkChannelTimeout is the maximum time to wait for the API response.
const probeLinkChannelTimeout = 30 * time.Second

// NewProbeLinkChannelCmd creates and returns the probe link-channel subcommand.
func NewProbeLinkChannelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-channel <probe-id> <channel-id>",
		Short: "Link a notification channel to a probe",
		Long: `Link a notification channel to a probe for alert notifications.

The probe can be specified by UUID or by name. If the name matches multiple
probes, you'll be prompted to use the UUID instead. The channel must still
be specified by UUID.

When a probe detects an issue, alerts will be sent to all linked channels.
A probe can have multiple channels linked, and a channel can be linked
to multiple probes.

Examples:
  # Link a channel to a probe by name
  stackeye probe link-channel "Production API" \
    660e8400-e29b-41d4-a716-446655440001

  # Link a channel to a probe by UUID
  stackeye probe link-channel 550e8400-e29b-41d4-a716-446655440000 \
    660e8400-e29b-41d4-a716-446655440001

Use 'stackeye probe get <probe-id>' to view currently linked channels.
Use 'stackeye channel list' to see available notification channels.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeLinkChannel(cmd, args[0], args[1])
		},
	}

	return cmd
}

// runProbeLinkChannel executes the probe link-channel command logic.
func runProbeLinkChannel(cmd *cobra.Command, probeIDArg, channelIDArg string) error {
	// Parse and validate channel UUID (channels are always referenced by UUID)
	channelID, err := uuid.Parse(channelIDArg)
	if err != nil {
		return fmt.Errorf("invalid channel ID %q: must be a valid UUID", channelIDArg)
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

	ctx, cancel := context.WithTimeout(cmd.Context(), probeLinkChannelTimeout)
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

	// Check if channel is already linked
	for _, existingID := range probe.AlertChannelIDs {
		if existingID == channelID {
			return fmt.Errorf("channel %q is already linked to probe %q", channel.Name, probe.Name)
		}
	}

	// Build update request with new channel added
	newChannelIDs := make([]uuid.UUID, len(probe.AlertChannelIDs)+1)
	copy(newChannelIDs, probe.AlertChannelIDs)
	newChannelIDs[len(probe.AlertChannelIDs)] = channelID
	req := &client.UpdateProbeRequest{
		AlertChannelIDs: newChannelIDs,
	}

	// Update the probe with the new channel list
	updatedProbe, err := client.UpdateProbe(ctx, apiClient, probeID, req)
	if err != nil {
		return fmt.Errorf("failed to link channel: %w", err)
	}

	// Print success message and updated probe
	fmt.Printf("Successfully linked channel %q to probe %q\n\n", channel.Name, probe.Name)
	return output.Print(updatedProbe)
}
