// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewAdminM2MKeyCmd creates and returns the admin m2m-key parent command.
// This command provides access to machine-to-machine key management features.
func NewAdminM2MKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "m2m-key",
		Short: "Manage machine-to-machine authentication keys",
		Long: `Manage machine-to-machine (M2M) authentication keys.

M2M keys are used for service-to-service authentication, allowing automated
systems and integrations to authenticate with the StackEye API without
user credentials.

Commands:
  list    List all M2M keys

Examples:
  # List all M2M keys
  stackeye admin m2m-key list

  # List M2M keys in JSON format
  stackeye admin m2m-key list -o json`,
		Aliases: []string{"m2m", "m2mkey"},
	}

	// Register subcommands
	cmd.AddCommand(NewAdminM2MKeyListCmd())

	return cmd
}
