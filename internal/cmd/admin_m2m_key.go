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
  list        List all M2M keys
  create      Create a new M2M key (regional or global)
  get         Get details of a specific M2M key
  rotate      Rotate an M2M key with grace period
  deactivate  Deactivate an M2M key immediately

Examples:
  # List all M2M keys
  stackeye admin m2m-key list

  # Create a global M2M key
  stackeye admin m2m-key create --global

  # Create a regional M2M key
  stackeye admin m2m-key create --region nyc3

  # Get M2M key details
  stackeye admin m2m-key get --id abc12345

  # Rotate an M2M key
  stackeye admin m2m-key rotate --id abc12345

  # Deactivate an M2M key
  stackeye admin m2m-key deactivate --id abc12345`,
		Aliases: []string{"m2m", "m2mkey"},
	}

	// Register subcommands
	cmd.AddCommand(NewAdminM2MKeyListCmd())
	cmd.AddCommand(NewAdminM2MKeyCreateCmd())
	cmd.AddCommand(NewAdminM2MKeyGetCmd())
	cmd.AddCommand(NewAdminM2MKeyRotateCmd())
	cmd.AddCommand(NewAdminM2MKeyDeactivateCmd())

	return cmd
}
