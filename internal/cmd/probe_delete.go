// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	cliinteractive "github.com/StackEye-IO/stackeye-cli/internal/interactive"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// probeDeleteTimeout is the maximum time to wait for a single delete API response.
const probeDeleteTimeout = 30 * time.Second

// probeDeleteFlags holds the flag values for the probe delete command.
type probeDeleteFlags struct {
	yes bool // Skip confirmation prompt
}

// NewProbeDeleteCmd creates and returns the probe delete subcommand.
func NewProbeDeleteCmd() *cobra.Command {
	flags := &probeDeleteFlags{}

	cmd := &cobra.Command{
		Use:               "delete <id> [id...]",
		Short:             "Delete one or more monitoring probes",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Delete one or more monitoring probes.

Probes can be specified by UUID or by name. If a name matches multiple probes,
you'll be prompted to use the UUID instead.

This permanently removes the probe(s) and all associated data including check history
and alert records. This action cannot be undone.

By default, the command will prompt for confirmation before deleting. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Delete a single probe by name
  stackeye probe delete "Production API"

  # Delete a single probe by UUID (with confirmation)
  stackeye probe delete 550e8400-e29b-41d4-a716-446655440000

  # Delete a probe without confirmation
  stackeye probe delete "Production API" --yes

  # Delete multiple probes at once (mix of names and UUIDs)
  stackeye probe delete "Production API" "Staging DB" 6ba7b810-9dad-11d1-80b4-00c04fd430c8

  # Delete multiple probes without confirmation (for scripting)
  stackeye probe delete --yes "Production API" "Staging DB"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeDelete(cmd.Context(), args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runProbeDelete executes the probe delete command logic.
func runProbeDelete(ctx context.Context, idArgs []string, flags *probeDeleteFlags) error {
	// Get authenticated API client first (needed for name resolution)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Resolve all probe identifiers (UUIDs or names) before prompting for confirmation
	probeIDs, err := ResolveProbeIDs(ctx, apiClient, idArgs)
	if err != nil {
		return err
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	message := "Are you sure you want to delete this probe?"
	if len(probeIDs) > 1 {
		message = fmt.Sprintf("Are you sure you want to delete %d probes?", len(probeIDs))
	}

	confirmed, err := cliinteractive.Confirm(message, cliinteractive.WithYesFlag(flags.yes))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Delete cancelled.")
		return nil
	}

	// Delete each probe
	var deleteErrors []error
	deletedCount := 0

	for _, probeID := range probeIDs {
		reqCtx, cancel := context.WithTimeout(ctx, probeDeleteTimeout)
		err := client.DeleteProbe(reqCtx, apiClient, probeID)
		cancel()

		if err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("failed to delete probe %s: %w", probeID, err))
			continue
		}

		deletedCount++
		fmt.Printf("Deleted probe %s\n", probeID)
	}

	// Report results
	if len(deleteErrors) > 0 {
		fmt.Printf("\nDeleted %d of %d probes.\n", deletedCount, len(probeIDs))
		for _, err := range deleteErrors {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Errorf("failed to delete %d probe(s)", len(deleteErrors))
	}

	if deletedCount > 1 {
		fmt.Printf("\nSuccessfully deleted %d probes.\n", deletedCount)
	}

	return nil
}
