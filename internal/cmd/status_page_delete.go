// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// statusPageDeleteTimeout is the maximum time to wait for a single delete API response.
const statusPageDeleteTimeout = 30 * time.Second

// statusPageDeleteFlags holds the flag values for the status-page delete command.
type statusPageDeleteFlags struct {
	yes bool // Skip confirmation prompt
}

// NewStatusPageDeleteCmd creates and returns the status-page delete subcommand.
func NewStatusPageDeleteCmd() *cobra.Command {
	flags := &statusPageDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a status page",
		Long: `Delete a status page by its ID.

This permanently removes the status page and all its configuration, including:
  - Associated probe display mappings
  - Custom domain configuration
  - Branding settings
  - All incident history displayed on the page

WARNING: This action cannot be undone. The status page URL will become unavailable
immediately after deletion.

By default, the command will prompt for confirmation before deleting. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Delete a status page (with confirmation)
  stackeye status-page delete 123

  # Delete a status page without confirmation
  stackeye status-page delete 123 --yes

  # Short form
  stackeye status-page delete 123 -y`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusPageDelete(cmd.Context(), args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runStatusPageDelete executes the status-page delete command logic.
func runStatusPageDelete(ctx context.Context, idArg string, flags *statusPageDeleteFlags) error {
	// Parse and validate status page ID
	id, err := strconv.ParseUint(idArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid status page ID %q: must be a positive integer", idArg)
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		dryrun.PrintAction("delete", "status page",
			"ID", idArg,
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Fetch status page to check if it exists and get details for display
	getCtx, cancelGet := context.WithTimeout(ctx, statusPageDeleteTimeout)
	statusPage, err := client.GetStatusPage(getCtx, apiClient, uint(id))
	cancelGet()
	if err != nil {
		return fmt.Errorf("failed to get status page: %w", err)
	}

	// Defensive check for nil status page
	if statusPage == nil {
		return fmt.Errorf("status page %d not found", id)
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	message := fmt.Sprintf("Are you sure you want to delete status page %d (%s)?", id, statusPage.Name)

	confirmed, err := cliinteractive.Confirm(message, cliinteractive.WithYesFlag(flags.yes))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Delete cancelled.")
		return nil
	}

	// Delete the status page
	deleteCtx, cancelDelete := context.WithTimeout(ctx, statusPageDeleteTimeout)
	err = client.DeleteStatusPage(deleteCtx, apiClient, uint(id))
	cancelDelete()

	if err != nil {
		return fmt.Errorf("failed to delete status page: %w", err)
	}

	fmt.Printf("Deleted status page %d (%s)\n", id, statusPage.Name)
	return nil
}
