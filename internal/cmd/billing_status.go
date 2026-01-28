// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// billingStatusTimeout is the maximum time to wait for the API response.
const billingStatusTimeout = 15 * time.Second

// NewBillingStatusCmd creates and returns the billing status command.
func NewBillingStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current billing status",
		Long: `Display lightweight billing status for your organization.

This command returns a quick billing status without making Stripe API calls,
making it faster than 'billing info'. Use this for quick status checks.

Shows:
  - Subscription status (active, trial, canceled, etc.)
  - Whether Stripe billing is configured

For detailed billing information including usage and payment details,
use 'stackeye billing info' instead.

Examples:
  # Show billing status
  stackeye billing status

  # Output as JSON for scripting
  stackeye billing status -o json

  # Output as YAML
  stackeye billing status -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBillingStatus(cmd.Context())
		},
	}

	return cmd
}

// runBillingStatus executes the billing status command logic.
func runBillingStatus(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get billing status with timeout
	reqCtx, cancel := context.WithTimeout(ctx, billingStatusTimeout)
	defer cancel()

	status, err := client.GetBillingStatus(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to get billing status: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(status)
		}
	}

	// Pretty print for table format (default)
	printBillingStatus(status)
	return nil
}

// printBillingStatus formats and prints the billing status in a human-friendly format.
func printBillingStatus(status *client.BillingStatus) {
	fmt.Println()
	fmt.Println("==================================")
	fmt.Println("         BILLING STATUS           ")
	fmt.Println("==================================")
	fmt.Println()
	fmt.Printf("  Status:   %s\n", formatBillingStatusValue(status.SubscriptionStatus))
	fmt.Printf("  Stripe:   %s\n", formatStripeStatus(status.HasStripeCustomer))
	fmt.Println()
	fmt.Println("----------------------------------")
	fmt.Println("  For detailed info:")
	fmt.Println("    stackeye billing info")
	fmt.Println()
	fmt.Println("  For usage metrics:")
	fmt.Println("    stackeye billing usage")
	fmt.Println("==================================")
	fmt.Println()
}

// formatBillingStatusValue formats the subscription status with an indicator.
func formatBillingStatusValue(status string) string {
	switch strings.ToLower(status) {
	case "active":
		return "[*] Active"
	case "trialing":
		return "[~] Trial"
	case "past_due":
		return "[!] Past Due"
	case "canceled", "cancelled":
		return "[ ] Canceled"
	case "incomplete":
		return "[.] Incomplete"
	case "incomplete_expired":
		return "[ ] Expired"
	case "unpaid":
		return "[!] Unpaid"
	case "paused":
		return "[-] Paused"
	case "none", "":
		return "[ ] None"
	default:
		return status
	}
}

// formatStripeStatus formats the Stripe customer status.
func formatStripeStatus(hasCustomer bool) string {
	if hasCustomer {
		return "Connected"
	}
	return "Not configured"
}
