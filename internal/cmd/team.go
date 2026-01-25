// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewTeamCmd creates and returns the team parent command.
// This command provides management of team members and invitations.
func NewTeamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage team members",
		Long: `Manage team members in your StackEye organization.

Team members are users who have access to your organization's resources
including probes, alerts, and notification channels. Each member has a role
that determines their permissions.

Roles:
  owner    Full control including billing, settings, and member management
  admin    Manage team members and all organization resources
  member   Create and manage probes, alerts, and channels
  viewer   Read-only access to all resources

Use 'stackeye team [command] --help' for information about available subcommands.`,
		Aliases: []string{"teams", "members"},
	}

	// Register subcommands
	cmd.AddCommand(NewTeamListCmd())
	cmd.AddCommand(NewTeamInviteCmd())
	cmd.AddCommand(NewTeamUpdateRoleCmd())

	return cmd
}
