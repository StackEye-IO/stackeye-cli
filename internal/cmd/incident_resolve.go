// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// incidentResolveTimeout is the maximum time to wait for the API response.
const incidentResolveTimeout = 30 * time.Second

// incidentResolveFlags holds the flag values for the incident resolve command.
type incidentResolveFlags struct {
	statusPageID uint
	incidentID   uint
	message      string
}

// NewIncidentResolveCmd creates and returns the incident resolve subcommand.
func NewIncidentResolveCmd() *cobra.Command {
	flags := &incidentResolveFlags{}

	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Resolve an incident on a status page",
		Long: `Resolve an incident to indicate that services have been restored.

Resolving an incident marks it as complete and records the resolution time.
Customers following your status page will see the incident as resolved.

Required Flags:
  --status-page-id   ID of the status page (required)
  --incident-id      ID of the incident to resolve (required)

Optional Flags:
  --message          Resolution message explaining what was fixed

Incident Status Workflow:
  investigating → identified → monitoring → resolved

When you resolve an incident:
  - Status is set to 'resolved'
  - Resolution timestamp is recorded
  - Customers see the incident as resolved on the status page
  - Optional: A resolution message can be added to explain the fix

Examples:
  # Resolve an incident
  stackeye incident resolve --status-page-id 123 --incident-id 456

  # Resolve with a resolution message
  stackeye incident resolve --status-page-id 123 --incident-id 456 \
    --message "Database connection pool increased. Issue resolved."

  # Output as JSON for scripting
  stackeye incident resolve --status-page-id 123 --incident-id 456 -o json

Note: Resolved incidents remain visible on the status page history. Use
'stackeye incident delete' to permanently remove an incident if needed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIncidentResolve(cmd.Context(), flags)
		},
	}

	// Required flags
	cmd.Flags().UintVar(&flags.statusPageID, "status-page-id", 0, "status page ID (required)")
	cmd.Flags().UintVar(&flags.incidentID, "incident-id", 0, "incident ID to resolve (required)")

	// Optional flags
	cmd.Flags().StringVar(&flags.message, "message", "", "resolution message explaining what was fixed")

	// Mark required flags
	_ = cmd.MarkFlagRequired("status-page-id")
	_ = cmd.MarkFlagRequired("incident-id")

	return cmd
}

// runIncidentResolve executes the incident resolve command logic.
func runIncidentResolve(ctx context.Context, flags *incidentResolveFlags) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, incidentResolveTimeout)
	defer cancel()

	// If message is provided, update the incident first to add the resolution message
	if flags.message != "" {
		updateReq := &client.UpdateIncidentRequest{
			Message: &flags.message,
		}
		_, err := client.UpdateIncident(reqCtx, apiClient, flags.statusPageID, flags.incidentID, updateReq)
		if err != nil {
			return fmt.Errorf("failed to add resolution message: %w", err)
		}
	}

	// Call SDK to resolve incident
	incident, err := client.ResolveIncident(reqCtx, apiClient, flags.statusPageID, flags.incidentID)
	if err != nil {
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	// Print the resolved incident using the table formatter
	return output.PrintIncident(*incident)
}
