// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewOrgCmd creates and returns the org parent command.
// This command provides management of organizations.
func NewOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage organizations",
		Long: `Manage organizations for your StackEye account.

Organizations are the top-level billing and access control entity in StackEye.
Each organization has its own subscription, team members, and resources (probes,
alerts, channels). Users can belong to multiple organizations and switch between
them.

Key Concepts:
  - Each organization has one subscription plan (Free, Starter, Pro, Team, Enterprise)
  - Resources (probes, alerts, channels) are scoped to organizations
  - Users can have different roles across organizations (owner, admin, member)
  - The CLI remembers your active organization in the current context

Use 'stackeye org [command] --help' for information about available subcommands.`,
		Aliases: []string{"orgs", "organization", "organizations"},
	}

	// Register subcommands
	cmd.AddCommand(NewOrgListCmd())
	cmd.AddCommand(NewOrgGetCmd())
	cmd.AddCommand(NewOrgSwitchCmd())

	return cmd
}
