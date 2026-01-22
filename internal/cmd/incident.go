// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewIncidentCmd creates and returns the incident parent command.
// This command provides management of status page incidents.
func NewIncidentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incident",
		Short: "Manage status page incidents",
		Long: `Manage incidents for public status pages.

Incidents communicate service disruptions to customers via your status pages.
Create incidents when issues occur, update them as you investigate, and resolve
them when services are restored.

Incident Status Workflow:
  investigating → identified → monitoring → resolved

Incident Impact Levels:
  none     - No impact to services
  minor    - Minor performance impact
  major    - Significant service degradation
  critical - Complete service outage

Examples:
  # List all incidents for a status page
  stackeye incident list --status-page-id 123

  # List only active incidents (not resolved)
  stackeye incident list --status-page-id 123 --status investigating

  # Create a new incident
  stackeye incident create --status-page-id 123 --title "API Degradation" --impact minor

  # Update an incident status
  stackeye incident update --status-page-id 123 --incident-id 456 --status identified

  # Resolve an incident
  stackeye incident resolve --status-page-id 123 --incident-id 456

  # Delete an incident
  stackeye incident delete --status-page-id 123 --incident-id 456

For more information about a specific command:
  stackeye incident [command] --help`,
		Aliases: []string{"inc", "incidents"},
	}

	// Register subcommands
	cmd.AddCommand(NewIncidentListCmd())
	cmd.AddCommand(NewIncidentCreateCmd())
	cmd.AddCommand(NewIncidentUpdateCmd())

	return cmd
}
