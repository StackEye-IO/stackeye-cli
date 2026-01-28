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
  status    Quick billing status check (fast, no Stripe API calls)
  info      Show detailed subscription and billing information
  usage     Show current resource usage against plan limits
  invoices  List billing invoices
  portal    Open Stripe billing portal

Examples:
  # Quick billing status check
  stackeye billing status

  # Show detailed billing info
  stackeye billing info

  # Show billing info as JSON
  stackeye billing info -o json

  # View current usage
  stackeye billing usage

  # List recent invoices
  stackeye billing invoices

  # Open Stripe billing portal
  stackeye billing portal`,
		Aliases: []string{"bill", "subscription"},
	}

	// Register subcommands
	cmd.AddCommand(NewBillingStatusCmd())
	cmd.AddCommand(NewBillingInfoCmd())
	cmd.AddCommand(NewBillingUsageCmd())
	cmd.AddCommand(NewBillingInvoicesCmd())
	cmd.AddCommand(NewBillingPortalCmd())

	return cmd
}
