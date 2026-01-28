// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// TestNewBillingInvoicesCmd_DownloadFlag verifies that the download flag is registered.
func TestNewBillingInvoicesCmd_DownloadFlag(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	flag := cmd.Flags().Lookup("download")
	if flag == nil {
		t.Fatal("expected --download flag to be registered")
	}

	if flag.DefValue != "false" {
		t.Errorf("expected download flag default to be 'false', got %q", flag.DefValue)
	}

	if flag.Usage == "" {
		t.Error("expected download flag to have a usage description")
	}
}

// TestNewBillingInvoicesCmd_OutputDirFlag verifies that the output-dir flag is registered.
func TestNewBillingInvoicesCmd_OutputDirFlag(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	flag := cmd.Flags().Lookup("output-dir")
	if flag == nil {
		t.Fatal("expected --output-dir flag to be registered")
	}

	if flag.DefValue != "." {
		t.Errorf("expected output-dir flag default to be '.', got %q", flag.DefValue)
	}

	if flag.Usage == "" {
		t.Error("expected output-dir flag to have a usage description")
	}
}

// TestNewBillingInvoicesCmd_Long_HasDownloadExamples verifies download examples in help.
func TestNewBillingInvoicesCmd_Long_HasDownloadExamples(t *testing.T) {
	cmd := NewBillingInvoicesCmd()

	long := cmd.Long

	// Should mention --download flag
	if !strings.Contains(long, "--download") {
		t.Error("expected Long description to mention --download flag")
	}

	// Should mention --output-dir flag
	if !strings.Contains(long, "--output-dir") {
		t.Error("expected Long description to mention --output-dir flag")
	}
}

// TestDownloadFile verifies HTTP file download functionality.
func TestDownloadFile(t *testing.T) {
	// Create a test server that serves PDF content
	pdfContent := []byte("%PDF-1.4 test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success.pdf":
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pdfContent)
		case "/error.pdf":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Save original validator and restore after test
	originalValidator := urlValidator
	defer func() { urlValidator = originalValidator }()

	// Override validator to allow test server URLs
	urlValidator = func(url string) bool {
		return strings.HasPrefix(url, server.URL) || isValidStripePDFURL(url)
	}

	// Create a temp directory for test files
	tempDir := t.TempDir()

	t.Run("successful download", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "test_success.pdf")

		err := downloadFile(context.Background(), server.URL+"/success.pdf", filePath)
		if err != nil {
			t.Fatalf("downloadFile failed: %v", err)
		}

		// Verify file was created with correct content
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read downloaded file: %v", err)
		}
		if string(data) != string(pdfContent) {
			t.Errorf("content mismatch: got %q, want %q", data, pdfContent)
		}
	})

	t.Run("HTTP error response", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "test_error.pdf")

		err := downloadFile(context.Background(), server.URL+"/error.pdf", filePath)
		if err == nil {
			t.Fatal("expected error for HTTP 404")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected 404 error, got: %v", err)
		}

		// Verify file was NOT created
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Error("file should not exist after failed download")
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "nonexistent", "dir", "test.pdf")

		err := downloadFile(context.Background(), server.URL+"/success.pdf", filePath)
		if err == nil {
			t.Fatal("expected error for non-existent directory")
		}
	})
}

// TestDownloadFile_WithStripeURL tests the full downloadFile function with mocked Stripe URL.
func TestDownloadFile_WithStripeURL(t *testing.T) {
	// Test URL validation rejects non-Stripe URLs
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.pdf")

	err := downloadFile(context.Background(), "https://evil.com/malware.pdf", filePath)
	if err == nil {
		t.Error("expected error for non-Stripe URL")
	}
	if !strings.Contains(err.Error(), "untrusted PDF URL domain") {
		t.Errorf("expected untrusted domain error, got: %v", err)
	}

	// Verify file was not created
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("file should not exist after rejected download")
	}
}

// TestDownloadInvoicePDFs_NoInvoicesWithPDF tests download with no PDF URLs.
func TestDownloadInvoicePDFs_NoInvoicesWithPDF(t *testing.T) {
	tempDir := t.TempDir()

	invoices := []client.Invoice{
		{ID: 1, InvoiceNumber: "INV-001", Status: "paid", Currency: "USD"},
		{ID: 2, InvoiceNumber: "INV-002", Status: "open", Currency: "USD"},
	}

	err := downloadInvoicePDFs(context.Background(), invoices, tempDir)
	if err != nil {
		t.Errorf("expected no error for invoices without PDFs, got: %v", err)
	}

	// Verify no files were created
	entries, _ := os.ReadDir(tempDir)
	if len(entries) != 0 {
		t.Errorf("expected no files, got %d", len(entries))
	}
}

// TestDownloadInvoicePDFs_SkipsExistingFiles tests that existing files are not overwritten.
func TestDownloadInvoicePDFs_SkipsExistingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create an existing file
	existingFile := filepath.Join(tempDir, "invoice-INV-001.pdf")
	if err := os.WriteFile(existingFile, []byte("existing content"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Use an invalid URL so download would fail if attempted
	badURL := "https://evil.com/bad.pdf"
	invoices := []client.Invoice{
		{ID: 1, InvoiceNumber: "INV-001", Status: "paid", Currency: "USD", PDFURL: &badURL},
	}

	// This should skip the existing file, not attempt download
	err := downloadInvoicePDFs(context.Background(), invoices, tempDir)
	// No error expected because file exists and was skipped
	if err != nil {
		t.Errorf("expected no error when skipping existing file, got: %v", err)
	}

	// Verify original content preserved
	data, _ := os.ReadFile(existingFile)
	if string(data) != "existing content" {
		t.Error("existing file was modified")
	}
}

// TestDownloadInvoicePDFs_InvalidURL tests handling of invalid URLs.
func TestDownloadInvoicePDFs_InvalidURL(t *testing.T) {
	tempDir := t.TempDir()

	badURL := "https://malicious.com/invoice.pdf"
	invoices := []client.Invoice{
		{ID: 1, InvoiceNumber: "INV-001", Status: "paid", Currency: "USD", PDFURL: &badURL},
	}

	err := downloadInvoicePDFs(context.Background(), invoices, tempDir)
	// Should return error for failed downloads
	if err == nil {
		t.Error("expected error for invalid URL download")
	}
	if !strings.Contains(err.Error(), "download(s) failed") {
		t.Errorf("expected download failed error, got: %v", err)
	}
}

// TestDownloadInvoicePDFs_SuccessfulDownload tests successful PDF downloads.
func TestDownloadInvoicePDFs_SuccessfulDownload(t *testing.T) {
	// Create a test server
	pdfContent := []byte("%PDF-1.4 test invoice content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pdfContent)
	}))
	defer server.Close()

	// Save original validator and restore after test
	originalValidator := urlValidator
	defer func() { urlValidator = originalValidator }()

	// Override validator to allow test server URLs
	urlValidator = func(url string) bool {
		return strings.HasPrefix(url, server.URL) || isValidStripePDFURL(url)
	}

	tempDir := t.TempDir()

	url1 := server.URL + "/invoice-001.pdf"
	url2 := server.URL + "/invoice-002.pdf"
	invoices := []client.Invoice{
		{ID: 1, InvoiceNumber: "INV-001", Status: "paid", Currency: "USD", PDFURL: &url1},
		{ID: 2, InvoiceNumber: "INV-002", Status: "open", Currency: "USD", PDFURL: &url2},
	}

	err := downloadInvoicePDFs(context.Background(), invoices, tempDir)
	if err != nil {
		t.Fatalf("downloadInvoicePDFs failed: %v", err)
	}

	// Verify files were created
	expectedFiles := []string{"invoice-INV-001.pdf", "invoice-INV-002.pdf"}
	for _, filename := range expectedFiles {
		filePath := filepath.Join(tempDir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("failed to read %s: %v", filename, err)
			continue
		}
		if string(data) != string(pdfContent) {
			t.Errorf("%s: content mismatch", filename)
		}
	}
}

// TestDownloadInvoicePDFs_PartialFailure tests handling of partial download failures.
func TestDownloadInvoicePDFs_PartialFailure(t *testing.T) {
	// Create a test server that fails on specific requests
	pdfContent := []byte("%PDF-1.4 test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "fail") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(pdfContent)
	}))
	defer server.Close()

	// Save original validator and restore after test
	originalValidator := urlValidator
	defer func() { urlValidator = originalValidator }()

	// Override validator to allow test server URLs
	urlValidator = func(url string) bool {
		return strings.HasPrefix(url, server.URL) || isValidStripePDFURL(url)
	}

	tempDir := t.TempDir()

	successURL := server.URL + "/invoice-success.pdf"
	failURL := server.URL + "/invoice-fail.pdf"
	invoices := []client.Invoice{
		{ID: 1, InvoiceNumber: "INV-001", Status: "paid", Currency: "USD", PDFURL: &successURL},
		{ID: 2, InvoiceNumber: "INV-002", Status: "open", Currency: "USD", PDFURL: &failURL},
	}

	err := downloadInvoicePDFs(context.Background(), invoices, tempDir)
	// Should return error due to partial failure
	if err == nil {
		t.Error("expected error for partial failure")
	}

	// First file should still be downloaded successfully
	data, err := os.ReadFile(filepath.Join(tempDir, "invoice-INV-001.pdf"))
	if err != nil {
		t.Errorf("successful download should still exist: %v", err)
	} else if string(data) != string(pdfContent) {
		t.Error("successful download content mismatch")
	}

	// Second file should NOT exist
	if _, err := os.Stat(filepath.Join(tempDir, "invoice-INV-002.pdf")); !os.IsNotExist(err) {
		t.Error("failed download file should not exist")
	}
}

// TestIsValidStripePDFURL verifies Stripe URL domain validation.
func TestIsValidStripePDFURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"invoice.stripe.com", "https://invoice.stripe.com/inv_abc123/pdf", true},
		{"pay.stripe.com", "https://pay.stripe.com/invoice/acct_xyz/invst_abc/pdf", true},
		{"files.stripe.com", "https://files.stripe.com/links/abc123/invoice.pdf", true},
		{"malicious domain", "https://evil.com/invoice.pdf", false},
		{"http not https", "http://invoice.stripe.com/inv_abc123/pdf", false},
		{"subdomain attack", "https://invoice.stripe.com.evil.com/pdf", false},
		{"empty string", "", false},
		{"local file", "file:///etc/passwd", false},
		{"data URL", "data:application/pdf;base64,abc123", false},
		{"stripe without https", "stripe.com/invoice", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStripePDFURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidStripePDFURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

// TestSanitizeFilename verifies filename sanitization.
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal filename", "invoice-001.pdf", "invoice-001.pdf"},
		{"with slash", "inv/001.pdf", "inv_001.pdf"},
		{"with backslash", "inv\\001.pdf", "inv_001.pdf"},
		{"with colon", "inv:001.pdf", "inv_001.pdf"},
		{"with asterisk", "inv*001.pdf", "inv_001.pdf"},
		{"with question mark", "inv?001.pdf", "inv_001.pdf"},
		{"with quotes", "inv\"001\".pdf", "inv_001_.pdf"},
		{"with angle brackets", "inv<001>.pdf", "inv_001_.pdf"},
		{"with pipe", "inv|001.pdf", "inv_001.pdf"},
		{"multiple invalid chars", "inv:/\\*?.pdf", "inv_____.pdf"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
