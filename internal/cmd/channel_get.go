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

// channelGetTimeout is the maximum time to wait for the API response.
const channelGetTimeout = 30 * time.Second

// NewChannelGetCmd creates and returns the channel get subcommand.
func NewChannelGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get details of a notification channel",
		Long: `Get detailed information about a specific notification channel.

Displays the full channel configuration including name, type, enabled status,
configuration details, and the number of probes linked to this channel.

Channel Types:
  email       Send notifications to email addresses
  slack       Post to Slack channels via incoming webhooks
  webhook     Send HTTP requests to custom endpoints
  pagerduty   Create incidents in PagerDuty
  discord     Post to Discord channels via webhooks
  teams       Post to Microsoft Teams channels
  sms         Send SMS text messages (requires SMS plan)

Examples:
  # Get channel details by ID
  stackeye channel get 550e8400-e29b-41d4-a716-446655440000

  # Output as JSON for scripting
  stackeye channel get 550e8400-e29b-41d4-a716-446655440000 -o json

  # Output as YAML
  stackeye channel get 550e8400-e29b-41d4-a716-446655440000 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChannelGet(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runChannelGet executes the channel get command logic.
func runChannelGet(ctx context.Context, idArg string) error {
	// Parse and validate UUID
	channelID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid channel ID %q: must be a valid UUID", idArg)
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get channel with timeout
	reqCtx, cancel := context.WithTimeout(ctx, channelGetTimeout)
	defer cancel()

	channel, err := client.GetChannel(reqCtx, apiClient, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Defensive check for nil channel
	if channel == nil {
		return fmt.Errorf("channel %s not found", channelID)
	}

	// Print the channel using the configured output format
	return output.Print(channel)
}
