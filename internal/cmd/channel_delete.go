// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/interactive"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// channelDeleteTimeout is the maximum time to wait for a single delete API response.
const channelDeleteTimeout = 30 * time.Second

// channelDeleteFlags holds the flag values for the channel delete command.
type channelDeleteFlags struct {
	yes bool // Skip confirmation prompt
}

// NewChannelDeleteCmd creates and returns the channel delete subcommand.
func NewChannelDeleteCmd() *cobra.Command {
	flags := &channelDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a notification channel",
		Long: `Delete a notification channel by its ID.

This permanently removes the channel configuration. Probes that were using this
channel for notifications will no longer send alerts through it.

If the channel is currently linked to any probes, you will be warned before
deletion. Consider updating those probes to use a different channel first.

By default, the command will prompt for confirmation before deleting. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Delete a channel (with confirmation)
  stackeye channel delete 550e8400-e29b-41d4-a716-446655440000

  # Delete a channel without confirmation
  stackeye channel delete 550e8400-e29b-41d4-a716-446655440000 --yes

  # Short form
  stackeye channel delete 550e8400-e29b-41d4-a716-446655440000 -y`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChannelDelete(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runChannelDelete executes the channel delete command logic.
func runChannelDelete(ctx context.Context, idArg string, flags *channelDeleteFlags) error {
	// Parse and validate UUID
	channelID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid channel ID %q: must be a valid UUID", idArg)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch channel to check if it exists and get probe count for warning
	getCtx, cancelGet := context.WithTimeout(ctx, channelDeleteTimeout)
	channel, err := client.GetChannel(getCtx, apiClient, channelID)
	cancelGet()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Defensive check for nil channel
	if channel == nil {
		return fmt.Errorf("channel %s not found", channelID)
	}

	// Warn if channel is linked to probes
	if channel.ProbeCount > 0 {
		fmt.Printf("Warning: This channel is linked to %d probe(s).\n", channel.ProbeCount)
		fmt.Println("Those probes will no longer send alerts through this channel after deletion.")
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	if !flags.yes && !GetNoInput() {
		message := fmt.Sprintf("Are you sure you want to delete channel %q?", channel.Name)

		confirmed, err := interactive.AskConfirm(&interactive.ConfirmPromptOptions{
			Message: message,
			Default: false,
		})
		if err != nil {
			if errors.Is(err, interactive.ErrPromptCancelled) {
				return fmt.Errorf("operation cancelled by user")
			}
			return fmt.Errorf("failed to prompt for confirmation: %w", err)
		}

		if !confirmed {
			fmt.Println("Delete cancelled.")
			return nil
		}
	}

	// Delete the channel
	deleteCtx, cancelDelete := context.WithTimeout(ctx, channelDeleteTimeout)
	err = client.DeleteChannel(deleteCtx, apiClient, channelID)
	cancelDelete()

	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	fmt.Printf("Deleted channel %s (%s)\n", channel.Name, channelID)
	return nil
}
