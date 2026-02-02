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

// incidentGetTimeout is the maximum time to wait for the API response.
const incidentGetTimeout = 30 * time.Second

// incidentGetFlags holds the flag values for the incident get command.
type incidentGetFlags struct {
	statusPageID uint
	incidentID   uint
}

// NewIncidentGetCmd creates and returns the incident get subcommand.
func NewIncidentGetCmd() *cobra.Command {
	flags := &incidentGetFlags{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a specific incident",
		Long: `Get detailed information about a specific incident by ID.

Displays the full incident details including title, message, status, impact level,
and all timestamps (created, updated, resolved).

Required Flags:
  --status-page-id   ID of the status page (required)
  --incident-id      ID of the incident to retrieve (required)

Incident Status Values:
  investigating - Initial investigation phase
  identified    - Root cause has been identified
  monitoring    - Fix deployed, monitoring for stability
  resolved      - Incident has been fully resolved

Impact Levels:
  none     - No impact to services (informational)
  minor    - Minor performance degradation
  major    - Significant service degradation
  critical - Complete service outage

Output Columns:
  ID        Incident ID
  TITLE     Incident title (truncated to 40 chars in table mode)
  STATUS    Current status with color coding
  IMPACT    Impact level with color coding
  CREATED   Creation timestamp

Wide Mode Columns (--output wide):
  UPDATED   Last update timestamp
  RESOLVED  Resolution timestamp (or - if not resolved)

Examples:
  # Get incident details
  stackeye incident get --status-page-id 123 --incident-id 456

  # Output as JSON for scripting
  stackeye incident get --status-page-id 123 --incident-id 456 -o json

  # Wide output with all columns
  stackeye incident get --status-page-id 123 --incident-id 456 -o wide

  # Use with jq for specific fields
  stackeye incident get --status-page-id 123 --incident-id 456 -o json | jq '.message'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIncidentGet(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().UintVar(&flags.statusPageID, "status-page-id", 0, "status page ID (required)")
	cmd.Flags().UintVar(&flags.incidentID, "incident-id", 0, "incident ID (required)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("status-page-id")
	_ = cmd.MarkFlagRequired("incident-id")

	return cmd
}

// runIncidentGet executes the incident get command logic.
func runIncidentGet(ctx context.Context, flags *incidentGetFlags) error {
	// Validate required fields
	if flags.statusPageID == 0 {
		return fmt.Errorf("--status-page-id is required")
	}

	if flags.incidentID == 0 {
		return fmt.Errorf("--incident-id is required")
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get incident with timeout
	reqCtx, cancel := context.WithTimeout(ctx, incidentGetTimeout)
	defer cancel()

	incident, err := client.GetIncident(reqCtx, apiClient, flags.statusPageID, flags.incidentID)
	if err != nil {
		return fmt.Errorf("failed to get incident: %w", err)
	}

	// Print the incident using the table formatter
	return output.PrintIncident(*incident)
}
