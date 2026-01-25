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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// billingInfoTimeout is the maximum time to wait for the API response.
const billingInfoTimeout = 30 * time.Second

// NewBillingInfoCmd creates and returns the billing info command.
func NewBillingInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show current subscription and billing information",
		Long: `Display detailed billing and subscription information for your organization.

Shows your current plan, subscription status, usage limits, and payment details.
For trial subscriptions, displays the trial end date and days remaining.

Output includes:
  - Plan name and subscription status
  - Monitor usage (current/limit)
  - Billing cycle and next billing date
  - Payment method on file
  - Trial information (if applicable)

Examples:
  # Show billing info
  stackeye billing info

  # Output as JSON for scripting
  stackeye billing info -o json

  # Output as YAML
  stackeye billing info -o yaml`,
		Aliases: []string{"show", "get", "status"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBillingInfo(cmd.Context())
		},
	}

	return cmd
}

// runBillingInfo executes the billing info command logic.
func runBillingInfo(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get billing info with timeout
	reqCtx, cancel := context.WithTimeout(ctx, billingInfoTimeout)
	defer cancel()

	info, err := client.GetBillingInfo(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to get billing info: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(info)
		}
	}

	// Pretty print for table format (default)
	printBillingInfo(info)
	return nil
}

// printBillingInfo formats and prints the billing info in a human-friendly format.
func printBillingInfo(info *client.BillingInfo) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    BILLING INFORMATION                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Subscription section
	fmt.Println("  ┌─────────────────────────────────┐")
	fmt.Println("  │         SUBSCRIPTION            │")
	fmt.Println("  ├─────────────────────────────────┤")
	fmt.Printf("  │  Plan:           %-14s │\n", formatPlanName(info.Plan))
	fmt.Printf("  │  Status:         %-14s │\n", formatStatus(info.Status))
	if info.BillingEmail != "" {
		fmt.Printf("  │  Billing Email:  %-14s │\n", truncateBillingField(info.BillingEmail, 14))
	}
	fmt.Println("  └─────────────────────────────────┘")
	fmt.Println()

	// Usage section
	fmt.Println("  ┌─────────────────────────────────┐")
	fmt.Println("  │            USAGE                │")
	fmt.Println("  ├─────────────────────────────────┤")
	usageStr := fmt.Sprintf("%d / %d", info.MonitorCount, info.MonitorLimit)
	usagePercent := float64(info.MonitorCount) / float64(info.MonitorLimit) * 100
	if info.MonitorLimit == 0 {
		usagePercent = 0
	}
	fmt.Printf("  │  Monitors:       %-14s │\n", usageStr)
	fmt.Printf("  │  Usage:          %5.1f%%         │\n", usagePercent)
	if info.CurrentPeriod != "" {
		fmt.Printf("  │  Period:         %-14s │\n", truncateBillingField(info.CurrentPeriod, 14))
	}
	fmt.Println("  └─────────────────────────────────┘")
	fmt.Println()

	// Trial info (if applicable)
	if info.TrialEndsAt != nil && *info.TrialEndsAt != "" {
		fmt.Println("  ┌─────────────────────────────────┐")
		fmt.Println("  │         TRIAL PERIOD            │")
		fmt.Println("  ├─────────────────────────────────┤")
		trialEnd := parseAndFormatDate(*info.TrialEndsAt)
		fmt.Printf("  │  Trial Ends:     %-14s │\n", trialEnd)
		daysRemaining := calculateDaysRemaining(*info.TrialEndsAt)
		if daysRemaining >= 0 {
			fmt.Printf("  │  Days Left:      %-14d │\n", daysRemaining)
		}
		fmt.Println("  └─────────────────────────────────┘")
		fmt.Println()
	}

	// Payment section
	fmt.Println("  ┌─────────────────────────────────┐")
	fmt.Println("  │           PAYMENT               │")
	fmt.Println("  ├─────────────────────────────────┤")
	if info.PaymentMethod != nil && *info.PaymentMethod != "" {
		fmt.Printf("  │  Method:         %-14s │\n", truncateBillingField(*info.PaymentMethod, 14))
	} else {
		fmt.Println("  │  Method:         Not on file    │")
	}
	if info.NextBillingAt != nil && *info.NextBillingAt != "" {
		nextBilling := parseAndFormatDate(*info.NextBillingAt)
		fmt.Printf("  │  Next Billing:   %-14s │\n", nextBilling)
	}
	if info.AmountCents != nil && *info.AmountCents > 0 {
		amount := formatCurrency(*info.AmountCents, info.Currency)
		fmt.Printf("  │  Amount:         %-14s │\n", amount)
	}
	if info.CancelAtPeriod {
		fmt.Println("  │  ⚠ Cancels at period end        │")
	}
	fmt.Println("  └─────────────────────────────────┘")
	fmt.Println()

	// Quick actions hint
	fmt.Println("  Quick Actions:")
	fmt.Println("    stackeye billing usage     - View detailed usage metrics")
	fmt.Println("    stackeye billing invoices  - View invoice history")
	fmt.Println()
}

// formatPlanName formats the plan name for display.
func formatPlanName(plan string) string {
	if plan == "" {
		return "None"
	}
	caser := cases.Title(language.English)
	return caser.String(strings.ToLower(plan))
}

// formatStatus formats the subscription status with an indicator.
func formatStatus(status string) string {
	switch strings.ToLower(status) {
	case "active":
		return "● Active"
	case "trialing":
		return "◐ Trial"
	case "past_due":
		return "⚠ Past Due"
	case "canceled", "cancelled":
		return "○ Canceled"
	case "incomplete":
		return "◌ Incomplete"
	case "incomplete_expired":
		return "○ Expired"
	case "unpaid":
		return "⚠ Unpaid"
	case "paused":
		return "⏸ Paused"
	default:
		if status == "" {
			return "No subscription"
		}
		return status
	}
}

// truncateBillingField truncates a string to fit in the billing display.
func truncateBillingField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// parseAndFormatDate parses an ISO date and returns a formatted date string.
func parseAndFormatDate(dateStr string) string {
	// Try parsing various date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("Jan 2, 2006")
		}
	}

	// Return original if parsing fails
	return truncateBillingField(dateStr, 14)
}

// calculateDaysRemaining calculates days until a date.
func calculateDaysRemaining(dateStr string) int {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			duration := time.Until(t)
			return int(duration.Hours() / 24)
		}
	}

	return -1
}

// formatCurrency formats cents to a currency string.
func formatCurrency(cents int, currency string) string {
	if currency == "" {
		currency = "USD"
	}
	currency = strings.ToUpper(currency)

	dollars := float64(cents) / 100.0

	// Use appropriate symbol
	symbol := "$"
	switch currency {
	case "EUR":
		symbol = "€"
	case "GBP":
		symbol = "£"
	case "JPY":
		symbol = "¥"
		dollars = float64(cents) // JPY doesn't use cents
	}

	return fmt.Sprintf("%s%.2f %s", symbol, dollars, currency)
}
