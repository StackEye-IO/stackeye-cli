// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// maintenanceDeleteTimeout is the maximum time to wait for a single delete API response.
const maintenanceDeleteTimeout = 30 * time.Second

// maintenanceDeleteFlags holds the flag values for the maintenance delete command.
type maintenanceDeleteFlags struct {
	yes bool // Skip confirmation prompt
}

// NewMaintenanceDeleteCmd creates and returns the maintenance delete subcommand.
func NewMaintenanceDeleteCmd() *cobra.Command {
	flags := &maintenanceDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a scheduled maintenance window",
		Long: `Delete a scheduled maintenance window by its ID.

This permanently removes the maintenance window configuration. Alerts that were
silenced by this maintenance window will resume normal notification behavior
after deletion.

Note: Deleting a maintenance window is different from letting it expire naturally.
Deletion removes it entirely, while expiration leaves a record in history. For
audit purposes, consider using 'stackeye mute expire' on the underlying mute if
you need to retain the record.

By default, the command will prompt for confirmation before deleting. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Delete a maintenance window (with confirmation)
  stackeye maintenance delete 550e8400-e29b-41d4-a716-446655440000

  # Delete a maintenance window without confirmation
  stackeye maintenance delete 550e8400-e29b-41d4-a716-446655440000 --yes

  # Short form
  stackeye maintenance delete 550e8400-e29b-41d4-a716-446655440000 -y`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMaintenanceDelete(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runMaintenanceDelete executes the maintenance delete command logic.
func runMaintenanceDelete(ctx context.Context, idArg string, flags *maintenanceDeleteFlags) error {
	// Parse and validate UUID
	maintenanceID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid maintenance window ID %q: must be a valid UUID", idArg)
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		dryrun.PrintAction("delete", "maintenance window",
			"ID", maintenanceID.String(),
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch maintenance window to check if it exists and get details for display
	// Maintenance windows are stored as mutes with IsMaintenanceWindow=true
	getCtx, cancelGet := context.WithTimeout(ctx, maintenanceDeleteTimeout)
	mute, err := client.GetMute(getCtx, apiClient, maintenanceID)
	cancelGet()
	if err != nil {
		return fmt.Errorf("failed to get maintenance window: %w", err)
	}

	// Defensive check for nil mute
	if mute == nil {
		return fmt.Errorf("maintenance window %s not found", maintenanceID)
	}

	// Verify this is actually a maintenance window, not a regular mute
	if !mute.IsMaintenanceWindow {
		return fmt.Errorf("ID %s is a mute, not a maintenance window; use 'stackeye mute delete' instead", maintenanceID)
	}

	// Get the maintenance name for display
	maintenanceName := "unnamed"
	if mute.MaintenanceName != nil && *mute.MaintenanceName != "" {
		maintenanceName = *mute.MaintenanceName
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	message := fmt.Sprintf("Are you sure you want to delete maintenance window %q (%s)?", maintenanceName, maintenanceID)

	confirmed, err := cliinteractive.Confirm(message, cliinteractive.WithYesFlag(flags.yes))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Delete cancelled.")
		return nil
	}

	// Delete the maintenance window (via DeleteMute since maintenance windows are mutes)
	deleteCtx, cancelDelete := context.WithTimeout(ctx, maintenanceDeleteTimeout)
	err = client.DeleteMute(deleteCtx, apiClient, maintenanceID)
	cancelDelete()

	if err != nil {
		return fmt.Errorf("failed to delete maintenance window: %w", err)
	}

	fmt.Printf("Deleted maintenance window %q (%s)\n", maintenanceName, maintenanceID)
	return nil
}
