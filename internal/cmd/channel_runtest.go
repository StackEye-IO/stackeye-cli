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

// channelTestTimeout is the maximum time to wait for the test API response.
// This includes time to send a test notification and receive the result.
const channelTestTimeout = 60 * time.Second

// ChannelTestResult wraps the test response with channel metadata for output formatting.
// This struct is exported to allow JSON/YAML serialization with proper field tags.
type ChannelTestResult struct {
	ChannelID      uuid.UUID `json:"channel_id" yaml:"channel_id"`
	ChannelName    string    `json:"channel_name" yaml:"channel_name"`
	ChannelType    string    `json:"channel_type" yaml:"channel_type"`
	Success        bool      `json:"success" yaml:"success"`
	Message        string    `json:"message" yaml:"message"`
	Error          *string   `json:"error,omitempty" yaml:"error,omitempty"`
	ResponseTimeMs int       `json:"response_time_ms" yaml:"response_time_ms"`
}

// NewChannelTestCmd creates and returns the channel test subcommand.
func NewChannelTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <id>",
		Short: "Send a test notification through a channel",
		Long: `Send a test notification through a notification channel to verify it's configured correctly.

This command sends a test notification message through the specified channel and
reports whether it was delivered successfully. Use this to verify that your channel
configuration (webhooks, email addresses, etc.) is working before relying on it
for actual alerts.

The test notification:
  - Uses the channel's current configuration
  - Sends a clearly marked test message (not a real alert)
  - Reports delivery success/failure and response time
  - Provides error details if delivery fails

This is useful for:
  - Verifying webhook URLs are correct and accessible
  - Confirming email addresses can receive notifications
  - Testing Slack/Discord/Teams integrations
  - Troubleshooting channel delivery issues

Examples:
  # Test a notification channel
  stackeye channel test 550e8400-e29b-41d4-a716-446655440000

  # Output as JSON for scripting
  stackeye channel test 550e8400-e29b-41d4-a716-446655440000 -o json

  # Using short form
  stackeye ch test 550e8400-e29b-41d4-a716-446655440000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChannelTest(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runChannelTest executes the channel test command logic.
func runChannelTest(ctx context.Context, idArg string) error {
	// Parse and validate UUID
	channelID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid channel ID %q: must be a valid UUID", idArg)
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		dryrun.PrintAction("send test notification to", "channel",
			"Channel ID", channelID.String(),
		)
		return nil
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout for the entire operation
	reqCtx, cancel := context.WithTimeout(ctx, channelTestTimeout)
	defer cancel()

	// First, fetch the channel to get its metadata
	fmt.Printf("Fetching channel %s...\n", channelID)
	channel, err := client.GetChannel(reqCtx, apiClient, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Defensive check for nil channel
	if channel == nil {
		return fmt.Errorf("channel %s not found", channelID)
	}

	// Send test notification
	fmt.Printf("Sending test notification to %q (%s)...\n", channel.Name, channel.Type)
	result, err := client.TestChannel(reqCtx, apiClient, channelID)
	if err != nil {
		return fmt.Errorf("failed to send test notification: %w", err)
	}

	// Create combined result with channel metadata
	testResult := &ChannelTestResult{
		ChannelID:      channel.ID,
		ChannelName:    channel.Name,
		ChannelType:    string(channel.Type),
		Success:        result.Success,
		Message:        result.Message,
		Error:          result.Error,
		ResponseTimeMs: result.ResponseTimeMs,
	}

	// Print using configured output format (supports json, yaml, table)
	return output.Print(testResult)
}
