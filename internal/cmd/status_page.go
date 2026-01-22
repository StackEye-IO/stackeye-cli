// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewStatusPageCmd creates and returns the status-page parent command.
// This command provides management of public status pages.
func NewStatusPageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status-page",
		Short: "Manage public status pages",
		Long: `Manage public status pages for your organization.

Status pages provide public-facing displays of system health. Each organization
can create multiple status pages with customizable branding and custom domains.

Key Features:
  - Custom subdomains (acme.stackeye.io) or custom domains (status.acme.com)
  - Component display from probes with optional display names
  - Automatic incident display derived from alerts
  - Uptime history visualization (90 days)
  - Branding customization (logo, favicon, colors, header/footer text)
  - Theme selection (light, dark, system)
  - Status badges for embedding in READMEs

Plan Limits:
  Free:       1 status page
  Starter:    2 status pages
  Pro:        5 status pages
  Team:       Unlimited
  Enterprise: Unlimited

Examples:
  # List all status pages
  stackeye status-page list

  # Get details of a specific status page
  stackeye status-page get 123

  # Create a new status page
  stackeye status-page create --name "Acme Status" --slug acme-status

  # Update a status page
  stackeye status-page update 123 --name "Acme Inc Status"

  # Delete a status page
  stackeye status-page delete 123

  # Add a probe to a status page
  stackeye status-page add-probe 123 --probe-id <uuid> --display-name "API"

  # Remove a probe from a status page
  stackeye status-page remove-probe 123 --probe-id <uuid>

  # Reorder probes on a status page
  stackeye status-page reorder-probes 123 --probe-id <uuid1>,<uuid2>,<uuid3>

  # Get aggregated status for a status page
  stackeye status-page get-status 123

  # Verify custom domain DNS configuration
  stackeye status-page domain-verify 123

For more information about a specific command:
  stackeye status-page [command] --help`,
		Aliases: []string{"sp", "statuspage"},
	}

	// Register subcommands
	cmd.AddCommand(NewStatusPageListCmd())
	cmd.AddCommand(NewStatusPageGetCmd())
	cmd.AddCommand(NewStatusPageCreateCmd())
	cmd.AddCommand(NewStatusPageUpdateCmd())
	cmd.AddCommand(NewStatusPageDeleteCmd())
	cmd.AddCommand(NewStatusPageAddProbeCmd())
	// cmd.AddCommand(NewStatusPageRemoveProbeCmd())
	// cmd.AddCommand(NewStatusPageReorderProbesCmd())
	// cmd.AddCommand(NewStatusPageGetStatusCmd())
	// cmd.AddCommand(NewStatusPageDomainVerifyCmd())

	return cmd
}
