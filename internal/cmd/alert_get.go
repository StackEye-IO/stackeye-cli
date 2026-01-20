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

// alertGetTimeout is the maximum time to wait for the API response.
const alertGetTimeout = 30 * time.Second

// alertGetFlags holds the flag values for the alert get command.
type alertGetFlags struct {
	timeline bool
}

// AlertGetResponse wraps the alert data with optional timeline for output.
type AlertGetResponse struct {
	Alert    *client.Alert               `json:"alert"`
	Timeline []client.AlertTimelineEvent `json:"timeline,omitempty"`
}

// NewAlertGetCmd creates and returns the alert get subcommand.
func NewAlertGetCmd() *cobra.Command {
	flags := &alertGetFlags{}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get details of a monitoring alert",
		Long: `Get detailed information about a specific monitoring alert.

Displays the full alert information including status, severity, triggered time,
associated probe details, and notification history.

Use the --timeline flag to include the event timeline showing all state
transitions and actions taken on the alert.

Alert States:
  active        Issue detected, not yet acknowledged
  acknowledged  Issue recognized, being investigated
  resolved      Issue fixed, alert closed

Alert Severities:
  critical      Service down or major outage
  warning       Degraded performance or minor issue
  info          Informational, no action required

Examples:
  # Get alert details by ID
  stackeye alert get 550e8400-e29b-41d4-a716-446655440000

  # Get alert with event timeline
  stackeye alert get 550e8400-e29b-41d4-a716-446655440000 --timeline

  # Output as JSON for scripting
  stackeye alert get 550e8400-e29b-41d4-a716-446655440000 -o json

  # Output as YAML
  stackeye alert get 550e8400-e29b-41d4-a716-446655440000 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlertGet(cmd.Context(), args[0], flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().BoolVar(&flags.timeline, "timeline", false, "include event timeline")

	return cmd
}

// runAlertGet executes the alert get command logic.
func runAlertGet(ctx context.Context, idArg string, flags *alertGetFlags) error {
	// Parse and validate UUID
	alertID, err := uuid.Parse(idArg)
	if err != nil {
		return fmt.Errorf("invalid alert ID %q: must be a valid UUID", idArg)
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get alert with timeout
	alertCtx, alertCancel := context.WithTimeout(ctx, alertGetTimeout)
	defer alertCancel()

	alert, err := client.GetAlert(alertCtx, apiClient, alertID)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	// Defensive check for nil alert
	if alert == nil {
		return fmt.Errorf("alert %s not found", alertID)
	}

	// Optionally fetch timeline if requested (separate timeout)
	var timeline []client.AlertTimelineEvent
	if flags.timeline {
		timelineCtx, timelineCancel := context.WithTimeout(ctx, alertGetTimeout)
		defer timelineCancel()

		timeline, err = client.GetAlertTimeline(timelineCtx, apiClient, alertID)
		if err != nil {
			return fmt.Errorf("failed to get alert timeline: %w", err)
		}
	}

	// Build response with optional timeline
	response := &AlertGetResponse{
		Alert:    alert,
		Timeline: timeline,
	}

	// Print the alert using the configured output format
	return output.Print(response)
}
