// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// billingInvoicesTimeout is the maximum time to wait for the API response.
const billingInvoicesTimeout = 30 * time.Second

// downloadTimeout is the maximum time to wait for a single PDF download.
const downloadTimeout = 60 * time.Second

// invoicesFlags holds flag values for the invoices command.
var invoicesFlags struct {
	limit     int
	download  bool
	outputDir string
}

// NewBillingInvoicesCmd creates and returns the billing invoices command.
func NewBillingInvoicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "List billing invoices",
		Long: `List billing invoices for your organization.

Shows a history of invoices including their status, amount, and payment date.
Use this to review past billing activity and download invoice PDFs.

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

  # Download all invoice PDFs to current directory
  stackeye billing invoices --download

  # Download invoice PDFs to a specific directory
  stackeye billing invoices --download --output-dir ~/invoices

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
	cmd.Flags().IntVar(&invoicesFlags.limit, "limit", 10, "Maximum number of invoices to retrieve (1-100)")
	cmd.Flags().BoolVar(&invoicesFlags.download, "download", false, "Download invoice PDFs to local directory")
	cmd.Flags().StringVar(&invoicesFlags.outputDir, "output-dir", ".", "Directory to save downloaded PDFs (default: current directory)")

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
	limit := max(invoicesFlags.limit, 1)
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

	// Handle download mode
	if invoicesFlags.download {
		return downloadInvoicePDFs(ctx, response.Invoices, invoicesFlags.outputDir)
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

// downloadInvoicePDFs downloads invoice PDFs to the specified directory.
func downloadInvoicePDFs(ctx context.Context, invoices []client.Invoice, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Count invoices with PDFs
	invoicesWithPDF := 0
	for _, inv := range invoices {
		if inv.PDFURL != nil && *inv.PDFURL != "" {
			invoicesWithPDF++
		}
	}

	if invoicesWithPDF == 0 {
		fmt.Println("No invoices with downloadable PDFs found.")
		return nil
	}

	fmt.Printf("Downloading %d invoice PDF(s) to %s\n\n", invoicesWithPDF, outputDir)

	downloaded := 0
	skipped := 0
	failed := 0

	for _, inv := range invoices {
		if inv.PDFURL == nil || *inv.PDFURL == "" {
			continue
		}

		// Generate filename: invoice-{number}.pdf
		filename := sanitizeFilename(fmt.Sprintf("invoice-%s.pdf", inv.InvoiceNumber))
		filePath := filepath.Join(outputDir, filename)

		// Check if file already exists
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("  ○ %s (already exists, skipping)\n", filename)
			skipped++
			continue
		}

		// Download the PDF
		fmt.Printf("  ● Downloading %s...", filename)
		if err := downloadFile(ctx, *inv.PDFURL, filePath); err != nil {
			fmt.Printf(" FAILED: %v\n", err)
			failed++
			continue
		}
		fmt.Println(" done")
		downloaded++
	}

	fmt.Println()
	fmt.Printf("Summary: %d downloaded, %d skipped, %d failed\n", downloaded, skipped, failed)

	if failed > 0 {
		return fmt.Errorf("%d download(s) failed", failed)
	}
	return nil
}

// downloadFile downloads a file from URL to the specified path.
func downloadFile(ctx context.Context, url, filePath string) error {
	// Validate URL domain for security - only allow known Stripe invoice domains
	if !urlValidator(url) {
		return fmt.Errorf("untrusted PDF URL domain: only Stripe invoice URLs are allowed")
	}

	// Create request with context and timeout
	reqCtx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create output file
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy response body to file
	if _, err := io.Copy(out, resp.Body); err != nil {
		// Clean up partial file on error
		os.Remove(filePath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// sanitizeFilename removes or replaces characters that are invalid in filenames.
func sanitizeFilename(name string) string {
	// Replace common invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// urlValidator is the function used to validate PDF URLs.
// It can be overridden in tests to allow localhost URLs.
var urlValidator = isValidStripePDFURL

// isValidStripePDFURL validates that the URL is from a trusted Stripe domain.
// This prevents potential SSRF attacks if the API were to return malicious URLs.
func isValidStripePDFURL(url string) bool {
	trustedPrefixes := []string{
		"https://invoice.stripe.com/",
		"https://pay.stripe.com/",
		"https://files.stripe.com/",
	}
	for _, prefix := range trustedPrefixes {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}
	return false
}
