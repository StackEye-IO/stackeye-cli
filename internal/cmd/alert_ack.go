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

// alertAckTimeout is the maximum time to wait for each API response.
const alertAckTimeout = 30 * time.Second

// alertAckFlags holds the flag values for the alert ack command.
type alertAckFlags struct {
	message string
}

// AlertAckResponse wraps the acknowledged alert data for output.
type AlertAckResponse struct {
	Alert   *client.Alert `json:"alert"`
	Message string        `json:"message,omitempty"`
}

// AlertAckBatchResponse wraps multiple acknowledged alerts for batch output.
type AlertAckBatchResponse struct {
	Acknowledged []*client.Alert `json:"acknowledged"`
	Failed       []AlertAckError `json:"failed,omitempty"`
	Total        int             `json:"total"`
	SuccessCount int             `json:"success_count"`
	FailedCount  int             `json:"failed_count"`
}

// AlertAckError represents a failed acknowledgment attempt.
type AlertAckError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// NewAlertAckCmd creates and returns the alert ack subcommand.
func NewAlertAckCmd() *cobra.Command {
	flags := &alertAckFlags{}

	cmd := &cobra.Command{
		Use:   "ack <id> [id...]",
		Short: "Acknowledge a monitoring alert",
		Long: `Acknowledge one or more monitoring alerts.

Acknowledging an alert indicates that the issue has been noticed and is being
investigated. This helps team coordination by showing which alerts are already
being addressed.

The acknowledgment is recorded with your user identity and timestamp. You can
optionally include a message providing context about the acknowledgment or
initial findings.

Alert States After Acknowledgment:
  acknowledged  Alert status changes from 'active' to 'acknowledged'
                Can still be resolved later when the issue is fixed

Examples:
  # Acknowledge a single alert
  stackeye alert ack 550e8400-e29b-41d4-a716-446655440000

  # Acknowledge with a note
  stackeye alert ack 550e8400-e29b-41d4-a716-446655440000 -m "Investigating high latency"

  # Acknowledge multiple alerts
  stackeye alert ack abc123... def456... ghi789...

  # Acknowledge multiple alerts with a shared note
  stackeye alert ack abc123... def456... -m "All related to network issue"

  # Output as JSON for scripting
  stackeye alert ack 550e8400-e29b-41d4-a716-446655440000 -o json`,
		Aliases: []string{"acknowledge"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAlertAck(cmd.Context(), args, flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().StringVarP(&flags.message, "message", "m", "", "note to include with acknowledgment")

	return cmd
}

// runAlertAck executes the alert ack command logic.
func runAlertAck(ctx context.Context, idArgs []string, flags *alertAckFlags) error {
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
	var req *client.AcknowledgeAlertRequest
	if flags.message != "" {
		req = &client.AcknowledgeAlertRequest{
			Note: &flags.message,
		}
	}

	// Single alert case - simpler output
	if len(alertIDs) == 1 {
		return runSingleAlertAck(ctx, apiClient, alertIDs[0], req, flags.message)
	}

	// Multiple alerts - batch output
	return runBatchAlertAck(ctx, apiClient, alertIDs, req)
}

// runSingleAlertAck acknowledges a single alert and outputs the result.
func runSingleAlertAck(ctx context.Context, apiClient *client.Client, alertID uuid.UUID, req *client.AcknowledgeAlertRequest, message string) error {
	// Call SDK to acknowledge alert with timeout
	reqCtx, cancel := context.WithTimeout(ctx, alertAckTimeout)
	defer cancel()

	alert, err := client.AcknowledgeAlert(reqCtx, apiClient, alertID, req)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	// Defensive check for nil alert
	if alert == nil {
		return fmt.Errorf("alert %s not found or already resolved", alertID)
	}

	// Build response
	response := &AlertAckResponse{
		Alert:   alert,
		Message: message,
	}

	// Print the acknowledged alert using the configured output format
	return output.Print(response)
}

// runBatchAlertAck acknowledges multiple alerts and outputs a summary.
func runBatchAlertAck(ctx context.Context, apiClient *client.Client, alertIDs []uuid.UUID, req *client.AcknowledgeAlertRequest) error {
	response := &AlertAckBatchResponse{
		Acknowledged: make([]*client.Alert, 0, len(alertIDs)),
		Failed:       make([]AlertAckError, 0),
		Total:        len(alertIDs),
	}

	// Process each alert
	for _, alertID := range alertIDs {
		// Call SDK to acknowledge alert with timeout
		reqCtx, cancel := context.WithTimeout(ctx, alertAckTimeout)
		alert, err := client.AcknowledgeAlert(reqCtx, apiClient, alertID, req)
		cancel()

		if err != nil {
			response.Failed = append(response.Failed, AlertAckError{
				ID:    alertID.String(),
				Error: err.Error(),
			})
			continue
		}

		if alert == nil {
			response.Failed = append(response.Failed, AlertAckError{
				ID:    alertID.String(),
				Error: "alert not found or already resolved",
			})
			continue
		}

		response.Acknowledged = append(response.Acknowledged, alert)
	}

	// Update counts
	response.SuccessCount = len(response.Acknowledged)
	response.FailedCount = len(response.Failed)

	// Print the batch result using the configured output format
	if err := output.Print(response); err != nil {
		return err
	}

	// Return error if any alerts failed
	if response.FailedCount > 0 {
		return fmt.Errorf("failed to acknowledge %d of %d alerts", response.FailedCount, response.Total)
	}

	return nil
}
