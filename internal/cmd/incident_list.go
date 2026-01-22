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

// incidentListTimeout is the maximum time to wait for the API response.
const incidentListTimeout = 30 * time.Second

// incidentListFlags holds the flag values for the incident list command.
type incidentListFlags struct {
	statusPageID uint
	page         int
	limit        int
	status       string
}

// NewIncidentListCmd creates and returns the incident list subcommand.
func NewIncidentListCmd() *cobra.Command {
	flags := &incidentListFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incidents for a status page",
		Long: `List all incidents for a specified status page.

Displays incidents with their ID, title, status, impact level, and timestamps.
Use the filter flags to narrow results by incident status.

Incident Status Workflow:
  investigating → identified → monitoring → resolved

Impact Levels:
  none     - No impact to services
  minor    - Minor performance impact
  major    - Significant service degradation
  critical - Complete service outage

Status Columns:
  ID        Incident ID
  TITLE     Incident title (truncated to 40 chars)
  STATUS    Current status (Investigating, Identified, Monitoring, Resolved)
  IMPACT    Impact level (None, Minor, Major, Critical)
  CREATED   Creation timestamp

Wide Mode Columns (--output wide):
  UPDATED   Last update timestamp
  RESOLVED  Resolution timestamp (or - if not resolved)

Examples:
  # List all incidents for a status page
  stackeye incident list --status-page-id 123

  # List only active incidents (investigating status)
  stackeye incident list --status-page-id 123 --status investigating

  # List resolved incidents
  stackeye incident list --status-page-id 123 --status resolved

  # Output as JSON for scripting
  stackeye incident list --status-page-id 123 -o json

  # Wide output with additional columns
  stackeye incident list --status-page-id 123 -o wide

  # Paginate through results
  stackeye incident list --status-page-id 123 --page 2 --limit 50`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIncidentList(cmd, flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().UintVar(&flags.statusPageID, "status-page-id", 0, "status page ID (required)")
	cmd.Flags().IntVar(&flags.page, "page", 1, "page number for pagination")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "results per page (max: 100)")
	cmd.Flags().StringVar(&flags.status, "status", "", "filter by status (investigating, identified, monitoring, resolved)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("status-page-id")

	return cmd
}

// runIncidentList executes the incident list command logic.
func runIncidentList(cmd *cobra.Command, flags *incidentListFlags) error {
	ctx := cmd.Context()

	// Validate all flags before making any API calls
	if flags.statusPageID == 0 {
		return fmt.Errorf("--status-page-id is required")
	}

	if flags.limit < 1 || flags.limit > 100 {
		return fmt.Errorf("invalid limit %d: must be between 1 and 100", flags.limit)
	}

	if flags.page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.page)
	}

	// Validate status filter if provided
	if flags.status != "" {
		validStatuses := map[string]bool{
			"investigating": true,
			"identified":    true,
			"monitoring":    true,
			"resolved":      true,
		}
		if !validStatuses[flags.status] {
			return fmt.Errorf("invalid status %q: must be investigating, identified, monitoring, or resolved", flags.status)
		}
	}

	// Get authenticated API client (after validation passes)
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options - SDK uses offset-based pagination
	offset := (flags.page - 1) * flags.limit
	opts := &client.ListIncidentsOptions{
		Limit:  flags.limit,
		Offset: offset,
	}

	// Add optional status filter
	if flags.status != "" {
		opts.Status = flags.status
	}

	// Call SDK to list incidents with timeout
	reqCtx, cancel := context.WithTimeout(ctx, incidentListTimeout)
	defer cancel()

	result, err := client.ListIncidents(reqCtx, apiClient, flags.statusPageID, opts)
	if err != nil {
		return fmt.Errorf("failed to list incidents: %w", err)
	}

	// Handle empty results
	if len(result.Incidents) == 0 {
		msg := "No incidents found"
		if flags.status != "" {
			msg = fmt.Sprintf("No incidents found with status '%s'", flags.status)
		}
		return output.PrintEmpty(msg)
	}

	// Print the incidents using the table formatter
	return output.PrintIncidents(result.Incidents)
}
