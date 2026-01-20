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

// probePauseTimeout is the maximum time to wait for a single pause API response.
const probePauseTimeout = 30 * time.Second

// probePauseFlags holds the flag values for the probe pause command.
type probePauseFlags struct {
	yes bool // Skip confirmation prompt
}

// NewProbePauseCmd creates and returns the probe pause subcommand.
func NewProbePauseCmd() *cobra.Command {
	flags := &probePauseFlags{}

	cmd := &cobra.Command{
		Use:   "pause <id> [id...]",
		Short: "Pause monitoring for one or more probes",
		Long: `Pause monitoring for one or more probes by their IDs.

Pausing a probe temporarily stops all monitoring checks without deleting the probe
configuration. This is useful during maintenance windows or when you need to
temporarily disable alerts.

Paused probes:
  - Stop executing scheduled checks
  - Do not trigger alerts
  - Retain all configuration and historical data
  - Can be resumed at any time with 'stackeye probe resume'

By default, the command will prompt for confirmation before pausing. Use --yes
to skip the confirmation prompt for scripting or automation.

Examples:
  # Pause a single probe (with confirmation)
  stackeye probe pause 550e8400-e29b-41d4-a716-446655440000

  # Pause a probe without confirmation
  stackeye probe pause 550e8400-e29b-41d4-a716-446655440000 --yes

  # Pause multiple probes at once
  stackeye probe pause 550e8400-e29b-41d4-a716-446655440000 6ba7b810-9dad-11d1-80b4-00c04fd430c8

  # Pause multiple probes without confirmation (for scripting)
  stackeye probe pause --yes 550e8400-e29b-41d4-a716-446655440000 6ba7b810-9dad-11d1-80b4-00c04fd430c8`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProbePause(cmd.Context(), args, flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.yes, "yes", "y", false, "skip confirmation prompt")

	return cmd
}

// runProbePause executes the probe pause command logic.
func runProbePause(ctx context.Context, idArgs []string, flags *probePauseFlags) error {
	// Parse and validate all UUIDs first before making any API calls
	probeIDs := make([]uuid.UUID, 0, len(idArgs))
	for _, idArg := range idArgs {
		probeID, err := uuid.Parse(idArg)
		if err != nil {
			return fmt.Errorf("invalid probe ID %q: must be a valid UUID", idArg)
		}
		probeIDs = append(probeIDs, probeID)
	}

	// Prompt for confirmation unless --yes flag is set or --no-input is enabled
	if !flags.yes && !GetNoInput() {
		message := "Are you sure you want to pause monitoring for this probe?"
		if len(probeIDs) > 1 {
			message = fmt.Sprintf("Are you sure you want to pause monitoring for %d probes?", len(probeIDs))
		}

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
			fmt.Println("Pause cancelled.")
			return nil
		}
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Pause each probe
	var pauseErrors []error
	pausedCount := 0

	for _, probeID := range probeIDs {
		reqCtx, cancel := context.WithTimeout(ctx, probePauseTimeout)
		probe, err := client.PauseProbe(reqCtx, apiClient, probeID)
		cancel()

		if err != nil {
			pauseErrors = append(pauseErrors, fmt.Errorf("failed to pause probe %s: %w", probeID, err))
			continue
		}

		pausedCount++
		fmt.Printf("Paused probe %s (%s) - status: %s\n", probeID, probe.Name, probe.Status)
	}

	// Report results
	if len(pauseErrors) > 0 {
		fmt.Printf("\nPaused %d of %d probes.\n", pausedCount, len(probeIDs))
		for _, err := range pauseErrors {
			fmt.Printf("Error: %v\n", err)
		}
		return fmt.Errorf("failed to pause %d probe(s)", len(pauseErrors))
	}

	if pausedCount > 1 {
		fmt.Printf("\nSuccessfully paused %d probes.\n", pausedCount)
	}

	return nil
}
