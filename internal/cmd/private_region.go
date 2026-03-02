// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewPrivateRegionCmd creates and returns the private-region parent command.
// This command provides access to private monitoring region lifecycle operations.
//
// Task: #10495 — Add CLI private-region lifecycle commands and tests (F-838)
func NewPrivateRegionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "private-region",
		Short: "Manage private monitoring regions",
		Long: `Manage private monitoring regions (org-owned self-hosted appliances).

Private regions run StackEye probe checks inside your own infrastructure,
such as corporate data centers or VPCs. Each region registers as a
self-hosted appliance and authenticates with a bootstrap key.

Commands:
  list    List all private regions
  get     Get details of a specific private region
  create  Create a new private region
  rotate  Rotate the bootstrap key for a private region
  revoke  Revoke a single bootstrap key

Examples:
  # List all private regions
  stackeye private-region list

  # Create a new private region
  stackeye private-region create --slug nyc-office --display-name "NYC Office" \
      --continent "North America" --country-code US

  # Get a private region by ID
  stackeye private-region get --id prv-nyc-office

  # Rotate bootstrap keys for a region
  stackeye private-region rotate --id prv-nyc-office

  # Revoke a single bootstrap key
  stackeye private-region revoke --region-id prv-nyc-office --key-id <uuid>`,
		Aliases: []string{"pr", "private-regions"},
	}

	// Register subcommands
	cmd.AddCommand(NewPrivateRegionListCmd())
	cmd.AddCommand(NewPrivateRegionGetCmd())
	cmd.AddCommand(NewPrivateRegionCreateCmd())
	cmd.AddCommand(NewPrivateRegionRotateCmd())
	cmd.AddCommand(NewPrivateRegionRevokeCmd())

	return cmd
}
