// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// incidentDeleteTimeout is the maximum time to wait for the API response.
const incidentDeleteTimeout = 30 * time.Second

// incidentDeleteFlags holds the flag values for the incident delete command.
type incidentDeleteFlags struct {
	statusPageID uint
	incidentID   uint
	force        bool
}

// NewIncidentDeleteCmd creates and returns the incident delete subcommand.
func NewIncidentDeleteCmd() *cobra.Command {
	flags := &incidentDeleteFlags{}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an incident from a status page",
		Long: `Permanently delete an incident from a status page.

WARNING: This action is irreversible. Once deleted, the incident and all its
history will be permanently removed from your status page.

Required Flags:
  --status-page-id   ID of the status page (required)
  --incident-id      ID of the incident to delete (required)

Optional Flags:
  --force            Skip confirmation prompt (useful for scripts)

When to Delete vs Resolve:
  - Resolve: Use when an incident was real and is now fixed. Keeps history.
  - Delete: Use when an incident was created by mistake or for testing.

Examples:
  # Delete an incident (will prompt for confirmation)
  stackeye incident delete --status-page-id 123 --incident-id 456

  # Delete an incident without confirmation (for scripts)
  stackeye incident delete --status-page-id 123 --incident-id 456 --force

  # Using short flags
  stackeye incident delete -s 123 -i 456 -f

Note: For incidents that have been resolved but should remain in history,
use 'stackeye incident resolve' instead.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIncidentDelete(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().UintVarP(&flags.statusPageID, "status-page-id", "s", 0, "status page ID (required)")
	cmd.Flags().UintVarP(&flags.incidentID, "incident-id", "i", 0, "incident ID to delete (required)")

	// Optional flags
	cmd.Flags().BoolVarP(&flags.force, "force", "f", false, "skip confirmation prompt")

	// Mark required flags
	_ = cmd.MarkFlagRequired("status-page-id")
	_ = cmd.MarkFlagRequired("incident-id")

	return cmd
}

// runIncidentDelete executes the incident delete command logic.
func runIncidentDelete(ctx context.Context, flags *incidentDeleteFlags) error {
	// Dry-run check: after flag parsing (cobra validates required flags), before API calls
	if GetDryRun() {
		dryrun.PrintAction("delete", "incident",
			"Status Page ID", fmt.Sprintf("%d", flags.statusPageID),
			"Incident ID", fmt.Sprintf("%d", flags.incidentID),
		)
		return nil
	}

	// Prompt for confirmation unless --force is specified
	if !flags.force {
		fmt.Printf("WARNING: This will permanently delete incident %d from status page %d.\n", flags.incidentID, flags.statusPageID)
		fmt.Print("This action is irreversible. Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			return fmt.Errorf("operation cancelled by user")
		}
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, incidentDeleteTimeout)
	defer cancel()

	// Call SDK to delete incident
	err = client.DeleteIncident(reqCtx, apiClient, flags.statusPageID, flags.incidentID)
	if err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	// Print success message (delete returns no data)
	fmt.Printf("Incident %d deleted successfully from status page %d\n", flags.incidentID, flags.statusPageID)

	return nil
}
