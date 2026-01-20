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

// alertResolveTimeout is the maximum time to wait for each API response.
const alertResolveTimeout = 30 * time.Second

// alertResolveFlags holds the flag values for the alert resolve command.
type alertResolveFlags struct {
	message string
}

// AlertResolveResponse wraps the resolved alert data for output.
type AlertResolveResponse struct {
	Alert   *client.Alert `json:"alert"`
	Message string        `json:"message,omitempty"`
}

// AlertResolveBatchResponse wraps multiple resolved alerts for batch output.
type AlertResolveBatchResponse struct {
	Resolved     []*client.Alert     `json:"resolved"`
	Failed       []AlertResolveError `json:"failed,omitempty"`
	Total        int                 `json:"total"`
	SuccessCount int                 `json:"success_count"`
	FailedCount  int                 `json:"failed_count"`
}

// AlertResolveError represents a failed resolution attempt.
type AlertResolveError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// NewAlertResolveCmd creates and returns the alert resolve subcommand.
func NewAlertResolveCmd() *cobra.Command {
	flags := &alertResolveFlags{}

	cmd := &cobra.Command{
		Use:   "resolve <id> [id...]",
		Short: "Resolve a monitoring alert",
		Long: `Resolve one or more monitoring alerts.

Resolving an alert marks the issue as fixed. This is typically done after the
underlying problem has been addressed and the monitored service has recovered.
The alert will be moved to 'resolved' status with a timestamp and the user who
resolved it recorded.

Note: Alerts can also be auto-resolved when probes detect recovery. Manual
resolution is useful when you've fixed the issue but the probe hasn't re-checked
yet, or for alerts that require manual intervention.

Alert States After Resolution:
  resolved    Alert status changes to 'resolved'
              Duration is calculated from triggered_at to resolved_at

Examples:
  # Resolve a single alert
  stackeye alert resolve 550e8400-e29b-41d4-a716-446655440000

  # Resolve with a note
  stackeye alert resolve 550e8400-e29b-41d4-a716-446655440000 -m "Fixed by restarting nginx"

  # Resolve multiple alerts
  stackeye alert resolve abc123... def456... ghi789...

  # Resolve multiple alerts with a shared note
  stackeye alert resolve abc123... def456... -m "All resolved by infrastructure upgrade"

  # Output as JSON for scripting
  stackeye alert resolve 550e8400-e29b-41d4-a716-446655440000 -o json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlertResolve(cmd.Context(), args, flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVarP(&flags.message, "message", "m", "", "note to include with resolution")

	return cmd
}

// runAlertResolve executes the alert resolve command logic.
func runAlertResolve(ctx context.Context, idArgs []string, flags *alertResolveFlags) error {
	// Parse and validate all UUIDs before making any API calls
	alertIDs := make([]uuid.UUID, 0, len(idArgs))
	for _, idArg := range idArgs {
		alertID, err := uuid.Parse(idArg)
		if err != nil {
			return fmt.Errorf("invalid alert ID %q: must be a valid UUID", idArg)
		}
		alertIDs = append(alertIDs, alertID)
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build request with optional note
	var req *client.ResolveAlertRequest
	if flags.message != "" {
		req = &client.ResolveAlertRequest{
			Note: &flags.message,
		}
	}

	// Single alert case - simpler output
	if len(alertIDs) == 1 {
		return runSingleAlertResolve(ctx, apiClient, alertIDs[0], req, flags.message)
	}

	// Multiple alerts - batch output
	return runBatchAlertResolve(ctx, apiClient, alertIDs, req)
}

// runSingleAlertResolve resolves a single alert and outputs the result.
func runSingleAlertResolve(ctx context.Context, apiClient *client.Client, alertID uuid.UUID, req *client.ResolveAlertRequest, message string) error {
	// Call SDK to resolve alert with timeout
	reqCtx, cancel := context.WithTimeout(ctx, alertResolveTimeout)
	defer cancel()

	alert, err := client.ResolveAlert(reqCtx, apiClient, alertID, req)
	if err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	// Defensive check for nil alert
	if alert == nil {
		return fmt.Errorf("alert %s not found or already resolved", alertID)
	}

	// Build response
	response := &AlertResolveResponse{
		Alert:   alert,
		Message: message,
	}

	// Print the resolved alert using the configured output format
	return output.Print(response)
}

// runBatchAlertResolve resolves multiple alerts and outputs a summary.
func runBatchAlertResolve(ctx context.Context, apiClient *client.Client, alertIDs []uuid.UUID, req *client.ResolveAlertRequest) error {
	response := &AlertResolveBatchResponse{
		Resolved: make([]*client.Alert, 0, len(alertIDs)),
		Failed:   make([]AlertResolveError, 0),
		Total:    len(alertIDs),
	}

	// Process each alert
	for _, alertID := range alertIDs {
		// Call SDK to resolve alert with timeout
		reqCtx, cancel := context.WithTimeout(ctx, alertResolveTimeout)
		alert, err := client.ResolveAlert(reqCtx, apiClient, alertID, req)
		cancel()

		if err != nil {
			response.Failed = append(response.Failed, AlertResolveError{
				ID:    alertID.String(),
				Error: err.Error(),
			})
			continue
		}

		if alert == nil {
			response.Failed = append(response.Failed, AlertResolveError{
				ID:    alertID.String(),
				Error: "alert not found or already resolved",
			})
			continue
		}

		response.Resolved = append(response.Resolved, alert)
	}

	// Update counts
	response.SuccessCount = len(response.Resolved)
	response.FailedCount = len(response.Failed)

	// Print the batch result using the configured output format
	if err := output.Print(response); err != nil {
		return err
	}

	// Return error if any alerts failed
	if response.FailedCount > 0 {
		return fmt.Errorf("failed to resolve %d of %d alerts", response.FailedCount, response.Total)
	}

	return nil
}
