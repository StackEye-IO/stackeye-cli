// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewBillingCmd creates and returns the billing parent command.
// This command provides access to billing and subscription information.
func NewBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Manage billing and subscription",
		Long: `Manage billing and subscription information for your StackEye organization.

View your current subscription plan, usage metrics, and billing details.
Access invoices and manage payment methods through the Stripe customer portal.

Commands:
  info      Show current subscription and billing information
  usage     Show current resource usage against plan limits
  invoices  List billing invoices

Examples:
  # Show current billing info
  stackeye billing info

  # Show billing info as JSON
  stackeye billing info -o json

  # View current usage
  stackeye billing usage

  # List recent invoices
  stackeye billing invoices`,
		Aliases: []string{"bill", "subscription"},
	}

	// Register subcommands
	cmd.AddCommand(NewBillingInfoCmd())
	cmd.AddCommand(NewBillingUsageCmd())

	return cmd
}
