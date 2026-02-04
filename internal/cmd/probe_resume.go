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
	"github.com/spf13/cobra"
)

// probeResumeTimeout is the maximum time to wait for a single resume API response.
const probeResumeTimeout = 30 * time.Second

// probeResumeFlags holds the flag values for the probe resume command.
type probeResumeFlags struct {
	yes bool // Skip confirmation prompt
}

// NewProbeResumeCmd creates and returns the probe resume subcommand.
func NewProbeResumeCmd() *cobra.Command {
	flags := &probeResumeFlags{}

	cmd := &cobra.Command{
		Use:               "resume <id> [id...]",
		Short:             "Resume monitoring for one or more paused probes",
		ValidArgsFunction: ProbeCompletion(),
		Long: `Resume monitoring for one or more paused probes.

Probes can be specified by UUID or by name. If a name matches multiple probes,
you'll be prompted to use the UUID instead.

Resuming a probe restarts all monitoring checks that were previously paused.
The probe will immediately begin executing scheduled checks again and can
trigger alerts based on the probe configuration.

Resumed probes:
  - Restart executing scheduled checks immediately
  - Can trigger alerts based on check results
  - Retain all configuration and historical data from before the pause
  - Status transitions from 'paused' to 'pending' until the first check completes

By default, the command will prompt for confirmation before resuming. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Resume a single probe by name
  stackeye probe resume "Production API"

  # Resume a single probe by UUID (with confirmation)
  stackeye probe resume 550e8400-e29b-41d4-a716-446655440000

  # Resume a probe without confirmation
  stackeye probe resume "Production API" --yes

  # Resume multiple probes at once (mix of names and UUIDs)
  stackeye probe resume "Production API" "Staging DB" 6ba7b810-9dad-11d1-80b4-00c04fd430c8

  # Resume multiple probes without confirmation (for scripting)
  stackeye probe resume --yes "Production API" "Staging DB"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbeResume(cmd.Context(), args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runProbeResume executes the probe resume command logic.
func runProbeResume(ctx context.Context, idArgs []string, flags *probeResumeFlags) error {
	// Dry-run check: print what would happen and exit without making API calls
	if GetDryRun() {
		dryrun.PrintBatchAction("resume", "probe", idArgs)
		return nil
	}

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
	message := "Are you sure you want to resume monitoring for this probe?"
	if len(probeIDs) > 1 {
		message = fmt.Sprintf("Are you sure you want to resume monitoring for %d probes?", len(probeIDs))
	}

	confirmed, err := cliinteractive.Confirm(message, cliinteractive.WithYesFlag(flags.yes))
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Resume cancelled.")
		return nil
	}

	// Resume each probe
	var resumeErrors []error
	resumedCount := 0

	for _, probeID := range probeIDs {
		reqCtx, cancel := context.WithTimeout(ctx, probeResumeTimeout)
		probe, err := client.ResumeProbe(reqCtx, apiClient, probeID)
		cancel()

		if err != nil {
			resumeErrors = append(resumeErrors, fmt.Errorf("failed to resume probe %s: %w", probeID, err))
			continue
		}

		resumedCount++
		fmt.Printf("Resumed probe %s (%s) - status: %s\n", probeID, probe.Name, probe.Status)
	}

	// Report results
	if len(resumeErrors) > 0 {
		fmt.Printf("\nResumed %d of %d probes.\n", resumedCount, len(probeIDs))
		for _, err := range resumeErrors {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Errorf("failed to resume %d probe(s)", len(resumeErrors))
	}

	if resumedCount > 1 {
		fmt.Printf("\nSuccessfully resumed %d probes.\n", resumedCount)
	}

	return nil
}
