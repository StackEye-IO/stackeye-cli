// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// muteExpireTimeout is the maximum time to wait for a single expire API response.
const muteExpireTimeout = 30 * time.Second

// muteExpireFlags holds the flag values for the mute expire command.
type muteExpireFlags struct {
	yes bool // Skip confirmation prompt
}

// NewMuteExpireCmd creates and returns the mute expire subcommand.
func NewMuteExpireCmd() *cobra.Command {
	flags := &muteExpireFlags{}

	cmd := &cobra.Command{
		Use:   "expire <id>",
		Short: "Immediately expire an active mute period",
		Long: `Immediately expire an active alert mute period by its ID.

This sets the mute's expiration time to now, ending its effect immediately while
preserving the mute record in history for audit purposes.

Note: Expiring a mute is different from deleting it. Expiration preserves the
record for compliance and audit trails, while deletion removes it entirely. For
normal operations, prefer using expire over delete.

By default, the command will prompt for confirmation before expiring. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Expire a mute (with confirmation)
  stackeye mute expire 550e8400-e29b-41d4-a716-446655440000

  # Expire a mute without confirmation
  stackeye mute expire 550e8400-e29b-41d4-a716-446655440000 --yes

  # Short form
  stackeye mute expire 550e8400-e29b-41d4-a716-446655440000 -y`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMuteExpire(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runMuteExpire executes the mute expire command logic.
func runMuteExpire(ctx context.Context, idArg string, flags *muteExpireFlags) error {
	// Parse and validate UUID
	muteID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid mute ID %q: must be a valid UUID", idArg)
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		dryrun.PrintAction("expire", "mute",
			"ID", muteID.String(),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch mute to check if it exists and get details for display
	getCtx, cancelGet := context.WithTimeout(ctx, muteExpireTimeout)
	mute, err := client.GetMute(getCtx, apiClient, muteID)
	cancelGet()
	if err != nil {
		return fmt.Errorf("failed to get mute: %w", err)
	}

	// Defensive check for nil mute
	if mute == nil {
		return fmt.Errorf("mute %s not found", muteID)
	}

	// Check if already expired
	if !mute.IsActive {
		return fmt.Errorf("mute %s is already expired", muteID)
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	message := fmt.Sprintf("Are you sure you want to expire mute %s (%s scope)?", muteID, mute.ScopeType)

	confirmed, err := cliinteractive.Confirm(message, cliinteractive.WithYesFlag(flags.yes))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Expire cancelled.")
		return nil
	}

	// Expire the mute
	expireCtx, cancelExpire := context.WithTimeout(ctx, muteExpireTimeout)
	expiredMute, err := client.ExpireMute(expireCtx, apiClient, muteID)
	cancelExpire()

	if err != nil {
		return fmt.Errorf("failed to expire mute: %w", err)
	}

	// Display the expired mute
	fmt.Printf("Expired mute %s (%s scope)\n\n", muteID, mute.ScopeType)
	if expiredMute != nil {
		return output.PrintMute(*expiredMute)
	}

	return nil
}
