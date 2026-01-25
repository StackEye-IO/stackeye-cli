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

// billingInvoicesTimeout is the maximum time to wait for the API response.
const billingInvoicesTimeout = 30 * time.Second

// invoicesLimit is the flag value for pagination.
var invoicesLimit int

// NewBillingInvoicesCmd creates and returns the billing invoices command.
func NewBillingInvoicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "List billing invoices",
		Long: `List billing invoices for your organization.

Shows a history of invoices including their status, amount, and payment date.
Use this to review past billing activity and access invoice PDFs.

Output includes:
  - Invoice number and date
  - Invoice status (paid, open, draft, void)
  - Total amount and currency
  - PDF URL for downloading invoices

Examples:
  # List recent invoices
  stackeye billing invoices

  # List more invoices
  stackeye billing invoices --limit 25

  # Output as JSON for scripting
  stackeye billing invoices -o json

  # Output as YAML
  stackeye billing invoices -o yaml`,
		Aliases: []string{"invoice", "inv"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBillingInvoices(cmd.Context())
		},
	}

	// Add flags
	cmd.Flags().IntVar(&invoicesLimit, "limit", 10, "Maximum number of invoices to retrieve (1-100)")

	return cmd
}

// runBillingInvoices executes the billing invoices command logic.
func runBillingInvoices(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Validate and clamp limit
	limit := max(invoicesLimit, 1)
	limit = min(limit, 100)

	// Call SDK to get invoices with timeout
	reqCtx, cancel := context.WithTimeout(ctx, billingInvoicesTimeout)
	defer cancel()

	opts := &client.ListInvoicesOptions{
		Limit: limit,
	}

	response, err := client.ListInvoices(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to get invoices: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	// Pretty print for table format (default)
	printInvoices(response)
	return nil
}

// printInvoices formats and prints the invoices in a human-friendly format.
func printInvoices(response *client.InvoiceListResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    BILLING INVOICES                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(response.Invoices) == 0 {
		fmt.Println("  No invoices found.")
		fmt.Println()
		return
	}

	// Summary line
	fmt.Printf("  Showing %d of %d invoice(s)\n", len(response.Invoices), response.Total)
	fmt.Println()

	// Print each invoice
	for i, inv := range response.Invoices {
		printInvoiceRow(inv, i == len(response.Invoices)-1)
	}

	// Pagination hint
	if response.HasMore {
		fmt.Println()
		fmt.Println("  More invoices available. Use --limit to retrieve more.")
	}

	// Quick actions hint
	fmt.Println()
	fmt.Println("  Quick Actions:")
	fmt.Println("    stackeye billing info   - View subscription details")
	fmt.Println("    stackeye billing usage  - View usage metrics")
	fmt.Println()
}

// printInvoiceRow prints a single invoice entry.
func printInvoiceRow(inv client.Invoice, isLast bool) {
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("  │  Invoice: %-46s │\n", truncateInvoiceField(inv.InvoiceNumber, 46))
	fmt.Println("  ├─────────────────────────────────────────────────────────┤")

	// Date
	createdDate := parseAndFormatDate(inv.CreatedAt)
	fmt.Printf("  │  Date:     %-45s │\n", createdDate)

	// Status with indicator
	statusStr := formatInvoiceStatus(inv.Status)
	fmt.Printf("  │  Status:   %-45s │\n", statusStr)

	// Amount
	amountStr := formatCurrency(int(inv.Total), inv.Currency)
	fmt.Printf("  │  Amount:   %-45s │\n", amountStr)

	// Paid date if applicable
	if inv.PaidAt != nil && *inv.PaidAt != "" {
		paidDate := parseAndFormatDate(*inv.PaidAt)
		fmt.Printf("  │  Paid:     %-45s │\n", paidDate)
	}

	// Period if available
	if inv.PeriodStart != nil && inv.PeriodEnd != nil && *inv.PeriodStart != "" && *inv.PeriodEnd != "" {
		periodStart := parseAndFormatDate(*inv.PeriodStart)
		periodEnd := parseAndFormatDate(*inv.PeriodEnd)
		periodStr := fmt.Sprintf("%s - %s", periodStart, periodEnd)
		fmt.Printf("  │  Period:   %-45s │\n", truncateInvoiceField(periodStr, 45))
	}

	// PDF link if available
	if inv.PDFURL != nil && *inv.PDFURL != "" {
		fmt.Printf("  │  PDF:      %-45s │\n", truncateInvoiceField(*inv.PDFURL, 45))
	}

	if isLast {
		fmt.Println("  └─────────────────────────────────────────────────────────┘")
	} else {
		fmt.Println("  └─────────────────────────────────────────────────────────┘")
		fmt.Println()
	}
}

// formatInvoiceStatus formats the invoice status with an indicator.
func formatInvoiceStatus(status string) string {
	switch strings.ToLower(status) {
	case "paid":
		return "● Paid"
	case "open":
		return "○ Open"
	case "draft":
		return "◌ Draft"
	case "void":
		return "○ Void"
	case "uncollectible":
		return "⚠ Uncollectible"
	default:
		if status == "" {
			return "Unknown"
		}
		return status
	}
}

// truncateInvoiceField truncates a string to fit in the invoice display.
func truncateInvoiceField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
