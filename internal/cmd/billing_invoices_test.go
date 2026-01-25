// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"slices"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewBillingInvoicesCmd verifies that the billing invoices command is properly constructed.
func TestNewBillingInvoicesCmd(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	if cmd.Use != "invoices" {
		t.Errorf("expected Use='invoices', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "List billing invoices" {
		t.Errorf("expected Short='List billing invoices', got %q", cmd.Short)
	}
}

// TestNewBillingInvoicesCmd_Aliases verifies that aliases are set correctly.
func TestNewBillingInvoicesCmd_Aliases(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	expectedAliases := []string{"invoice", "inv"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, alias := range expectedAliases {
		if !slices.Contains(cmd.Aliases, alias) {
			t.Errorf("expected alias %q not found", alias)
		}
	}
}

// TestNewBillingInvoicesCmd_Long verifies the Long description contains key information.
func TestNewBillingInvoicesCmd_Long(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"invoice number",
		"status",
		"amount",
		"PDF",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye billing invoices") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention output formats
	if !strings.Contains(long, "json") {
		t.Error("expected Long description to mention JSON output option")
	}

	// Should mention limit flag
	if !strings.Contains(long, "--limit") {
		t.Error("expected Long description to mention --limit flag")
	}
}

// TestNewBillingInvoicesCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewBillingInvoicesCmd_RunEIsSet(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewBillingInvoicesCmd_LimitFlag verifies that the limit flag is registered.
func TestNewBillingInvoicesCmd_LimitFlag(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("expected --limit flag to be registered")
	}

	if flag.DefValue != "10" {
		t.Errorf("expected limit flag default to be '10', got %q", flag.DefValue)
	}

	// Check flag usage contains description
	if flag.Usage == "" {
		t.Error("expected limit flag to have a usage description")
	}
}

// TestFormatInvoiceStatus verifies invoice status formatting.
func TestFormatInvoiceStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		contains string
	}{
		{"paid", "paid", "Paid"},
		{"open", "open", "Open"},
		{"draft", "draft", "Draft"},
		{"void", "void", "Void"},
		{"uncollectible", "uncollectible", "Uncollectible"},
		{"uppercase paid", "PAID", "Paid"},
		{"mixed case", "Paid", "Paid"},
		{"empty", "", "Unknown"},
		{"unknown status", "custom_status", "custom_status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInvoiceStatus(tt.status)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatInvoiceStatus(%q) = %q, expected to contain %q", tt.status, result, tt.contains)
			}
		})
	}
}

// TestFormatInvoiceStatus_HasIndicators verifies status formatting has visual indicators.
func TestFormatInvoiceStatus_HasIndicators(t *testing.T) {
	tests := []struct {
		status    string
		indicator string
	}{
		{"paid", "●"},
		{"open", "○"},
		{"draft", "◌"},
		{"void", "○"},
		{"uncollectible", "⚠"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := formatInvoiceStatus(tt.status)
			if !strings.Contains(result, tt.indicator) {
				t.Errorf("formatInvoiceStatus(%q) = %q, expected indicator %q", tt.status, result, tt.indicator)
			}
		})
	}
}

// TestTruncateInvoiceField verifies invoice field truncation.
func TestTruncateInvoiceField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"longer truncated", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "hel"},
		{"empty string", "", 10, ""},
		{"single char max", "hello", 1, "h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateInvoiceField(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateInvoiceField(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestPrintInvoices_DoesNotPanic verifies that printInvoices doesn't panic with various inputs.
func TestPrintInvoices_DoesNotPanic(t *testing.T) {
	pdfURL := "https://example.com/invoice.pdf"
	periodStart := "2026-01-01T00:00:00Z"
	periodEnd := "2026-01-31T23:59:59Z"
	paidAt := "2026-01-15T10:30:00Z"

	tests := []struct {
		name     string
		response *client.InvoiceListResponse
	}{
		{
			name: "empty invoices",
			response: &client.InvoiceListResponse{
				Invoices: []client.Invoice{},
				Total:    0,
				HasMore:  false,
			},
		},
		{
			name: "single invoice minimal",
			response: &client.InvoiceListResponse{
				Invoices: []client.Invoice{
					{
						ID:            1,
						InvoiceNumber: "INV-001",
						Status:        "paid",
						Total:         1000,
						Currency:      "USD",
						CreatedAt:     "2026-01-01T00:00:00Z",
					},
				},
				Total:   1,
				HasMore: false,
			},
		},
		{
			name: "invoice with all fields",
			response: &client.InvoiceListResponse{
				Invoices: []client.Invoice{
					{
						ID:            1,
						InvoiceNumber: "INV-001",
						Status:        "paid",
						Subtotal:      900,
						Tax:           100,
						Total:         1000,
						AmountPaid:    1000,
						AmountDue:     0,
						Currency:      "USD",
						PeriodStart:   &periodStart,
						PeriodEnd:     &periodEnd,
						PaidAt:        &paidAt,
						PDFURL:        &pdfURL,
						CreatedAt:     "2026-01-01T00:00:00Z",
					},
				},
				Total:   1,
				HasMore: false,
			},
		},
		{
			name: "multiple invoices",
			response: &client.InvoiceListResponse{
				Invoices: []client.Invoice{
					{
						ID:            1,
						InvoiceNumber: "INV-001",
						Status:        "paid",
						Total:         1000,
						Currency:      "USD",
						CreatedAt:     "2026-01-01T00:00:00Z",
					},
					{
						ID:            2,
						InvoiceNumber: "INV-002",
						Status:        "open",
						Total:         2500,
						Currency:      "EUR",
						CreatedAt:     "2026-02-01T00:00:00Z",
					},
					{
						ID:            3,
						InvoiceNumber: "INV-003",
						Status:        "draft",
						Total:         500,
						Currency:      "GBP",
						CreatedAt:     "2026-03-01T00:00:00Z",
					},
				},
				Total:   3,
				HasMore: false,
			},
		},
		{
			name: "with pagination",
			response: &client.InvoiceListResponse{
				Invoices: []client.Invoice{
					{
						ID:            1,
						InvoiceNumber: "INV-001",
						Status:        "paid",
						Total:         1000,
						Currency:      "USD",
						CreatedAt:     "2026-01-01T00:00:00Z",
					},
				},
				Total:   50,
				HasMore: true,
			},
		},
		{
			name: "various statuses",
			response: &client.InvoiceListResponse{
				Invoices: []client.Invoice{
					{ID: 1, InvoiceNumber: "INV-001", Status: "paid", Total: 1000, Currency: "USD", CreatedAt: "2026-01-01T00:00:00Z"},
					{ID: 2, InvoiceNumber: "INV-002", Status: "open", Total: 1000, Currency: "USD", CreatedAt: "2026-01-01T00:00:00Z"},
					{ID: 3, InvoiceNumber: "INV-003", Status: "draft", Total: 1000, Currency: "USD", CreatedAt: "2026-01-01T00:00:00Z"},
					{ID: 4, InvoiceNumber: "INV-004", Status: "void", Total: 1000, Currency: "USD", CreatedAt: "2026-01-01T00:00:00Z"},
					{ID: 5, InvoiceNumber: "INV-005", Status: "uncollectible", Total: 1000, Currency: "USD", CreatedAt: "2026-01-01T00:00:00Z"},
				},
				Total:   5,
				HasMore: false,
			},
		},
		{
			name: "long invoice number",
			response: &client.InvoiceListResponse{
				Invoices: []client.Invoice{
					{
						ID:            1,
						InvoiceNumber: "INV-VERY-LONG-INVOICE-NUMBER-THAT-EXCEEDS-NORMAL-LENGTH-12345678901234567890",
						Status:        "paid",
						Total:         1000,
						Currency:      "USD",
						CreatedAt:     "2026-01-01T00:00:00Z",
					},
				},
				Total:   1,
				HasMore: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printInvoices panicked: %v", r)
				}
			}()
			printInvoices(tt.response)
		})
	}
}

// TestNewBillingCmd_HasInvoicesSubcommand verifies that invoices subcommand is registered.
func TestNewBillingCmd_HasInvoicesSubcommand(t *testing.T) {
	cmd := NewBillingCmd()

	subcommands := cmd.Commands()
	if len(subcommands) < 3 {
		t.Error("expected billing command to have at least 3 subcommands (info, usage, and invoices)")
	}

	// Verify invoices subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "invoices" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'invoices' subcommand to be registered")
	}
}
