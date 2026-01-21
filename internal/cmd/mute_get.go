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

// muteGetTimeout is the maximum time to wait for the API response.
const muteGetTimeout = 30 * time.Second

// NewMuteGetCmd creates and returns the mute get subcommand.
func NewMuteGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get details of an alert mute period",
		Long: `Get detailed information about a specific alert mute period.

Displays the full mute information including scope, target, duration,
expiration time, status, and associated metadata.

Mute Scopes:
  organization  Mute affecting all alerts in the organization
  probe         Mute affecting a specific probe
  channel       Mute affecting a specific notification channel
  alert_type    Mute affecting alerts of a specific type

Status Values:
  ACTIVE        Mute is currently in effect
  EXPIRED       Mute has ended

Examples:
  # Get mute details by ID
  stackeye mute get 550e8400-e29b-41d4-a716-446655440000

  # Output as JSON for scripting
  stackeye mute get 550e8400-e29b-41d4-a716-446655440000 -o json

  # Output as YAML
  stackeye mute get 550e8400-e29b-41d4-a716-446655440000 -o yaml

  # Wide output with additional columns
  stackeye mute get 550e8400-e29b-41d4-a716-446655440000 -o wide`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMuteGet(cmd.Context(), args[0])
		},
	}

	return cmd
}

// runMuteGet executes the mute get command logic.
func runMuteGet(ctx context.Context, idArg string) error {
	// Parse and validate UUID
	muteID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid mute ID %q: must be a valid UUID", idArg)
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get mute with timeout
	reqCtx, cancel := context.WithTimeout(ctx, muteGetTimeout)
	defer cancel()

	mute, err := client.GetMute(reqCtx, apiClient, muteID)
	if err != nil {
		return fmt.Errorf("failed to get mute: %w", err)
	}

	// Defensive check for nil mute
	if mute == nil {
		return fmt.Errorf("mute %s not found", muteID)
	}

	// Print the mute using the configured output format
	return output.PrintMute(*mute)
}
