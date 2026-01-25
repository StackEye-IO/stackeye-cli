// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"slices"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewBillingInfoCmd verifies that the billing info command is properly constructed.
func TestNewBillingInfoCmd(t *testing.T) {
	cmd := NewBillingInfoCmd()

	if cmd.Use != "info" {
		t.Errorf("expected Use='info', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Show current subscription and billing information" {
		t.Errorf("expected Short='Show current subscription and billing information', got %q", cmd.Short)
	}
}

// TestNewBillingInfoCmd_Aliases verifies that aliases are set correctly.
func TestNewBillingInfoCmd_Aliases(t *testing.T) {
	cmd := NewBillingInfoCmd()

	expectedAliases := []string{"show", "get", "status"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, alias := range expectedAliases {
		if !slices.Contains(cmd.Aliases, alias) {
			t.Errorf("expected alias %q not found", alias)
		}
	}
}

// TestNewBillingInfoCmd_Long verifies the Long description contains key information.
func TestNewBillingInfoCmd_Long(t *testing.T) {
	cmd := NewBillingInfoCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"Plan name",
		"Monitor usage",
		"Billing cycle",
		"Payment method",
		"Trial information",
	}
	for _, feature := range features {
		if !strings.Contains(long, feature) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye billing info") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention output formats
	if !strings.Contains(long, "json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

// TestNewBillingInfoCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewBillingInfoCmd_RunEIsSet(t *testing.T) {
	cmd := NewBillingInfoCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewBillingCmd verifies that the billing parent command is properly constructed.
func TestNewBillingCmd(t *testing.T) {
	cmd := NewBillingCmd()

	if cmd.Use != "billing" {
		t.Errorf("expected Use='billing', got %q", cmd.Use)
	}

	if cmd.Short != "Manage billing and subscription" {
		t.Errorf("expected Short='Manage billing and subscription', got %q", cmd.Short)
	}
}

// TestNewBillingCmd_Aliases verifies that aliases are set correctly.
func TestNewBillingCmd_Aliases(t *testing.T) {
	cmd := NewBillingCmd()

	expectedAliases := []string{"bill", "subscription"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, alias := range expectedAliases {
		if !slices.Contains(cmd.Aliases, alias) {
			t.Errorf("expected alias %q not found", alias)
		}
	}
}

// TestNewBillingCmd_HasSubcommands verifies that subcommands are registered.
func TestNewBillingCmd_HasSubcommands(t *testing.T) {
	cmd := NewBillingCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected billing command to have subcommands")
	}

	// Verify info subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "info" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'info' subcommand to be registered")
	}
}

// TestFormatPlanName verifies plan name formatting.
func TestFormatPlanName(t *testing.T) {
	tests := []struct {
		name     string
		plan     string
		expected string
	}{
		{"empty plan", "", "None"},
		{"lowercase plan", "starter", "Starter"},
		{"uppercase plan", "PRO", "Pro"},
		{"mixed case plan", "EnTeRpRiSe", "Enterprise"},
		{"team plan", "team", "Team"},
		{"free plan", "free", "Free"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPlanName(tt.plan)
			if result != tt.expected {
				t.Errorf("formatPlanName(%q) = %q, want %q", tt.plan, result, tt.expected)
			}
		})
	}
}

// TestFormatStatus verifies subscription status formatting.
func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"active", "active", "● Active"},
		{"trialing", "trialing", "◐ Trial"},
		{"past due", "past_due", "⚠ Past Due"},
		{"canceled lowercase", "canceled", "○ Canceled"},
		{"cancelled UK spelling", "cancelled", "○ Canceled"},
		{"incomplete", "incomplete", "◌ Incomplete"},
		{"incomplete_expired", "incomplete_expired", "○ Expired"},
		{"unpaid", "unpaid", "⚠ Unpaid"},
		{"paused", "paused", "⏸ Paused"},
		{"empty status", "", "No subscription"},
		{"unknown status", "pending", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStatus(tt.status)
			if result != tt.expected {
				t.Errorf("formatStatus(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

// TestTruncateBillingField verifies string truncation.
func TestTruncateBillingField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 2, "he"},
		{"max is 3", "hello", 3, "hel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateBillingField(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateBillingField(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestParseAndFormatDate verifies date parsing and formatting.
func TestParseAndFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"RFC3339 format", "2026-02-15T12:00:00Z", "Feb 15, 2026"},
		{"date only", "2026-03-01", "Mar 1, 2026"},
		{"RFC3339 with timezone", "2026-01-01T00:00:00+00:00", "Jan 1, 2026"},
		{"invalid format", "invalid-date", "invalid-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAndFormatDate(tt.input)
			if result != tt.expected {
				t.Errorf("parseAndFormatDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCalculateDaysRemaining verifies days remaining calculation.
func TestCalculateDaysRemaining(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"invalid date returns -1", "not-a-date", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDaysRemaining(tt.input)
			if result != tt.expected {
				t.Errorf("calculateDaysRemaining(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFormatCurrency verifies currency formatting.
func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name     string
		cents    int
		currency string
		expected string
	}{
		{"USD dollars", 1200, "USD", "$12.00 USD"},
		{"USD default", 500, "", "$5.00 USD"},
		{"EUR", 2500, "EUR", "€25.00 EUR"},
		{"GBP", 1599, "GBP", "£15.99 GBP"},
		{"JPY no decimals", 1000, "JPY", "¥1000.00 JPY"},
		{"lowercase currency", 100, "usd", "$1.00 USD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCurrency(tt.cents, tt.currency)
			if result != tt.expected {
				t.Errorf("formatCurrency(%d, %q) = %q, want %q", tt.cents, tt.currency, result, tt.expected)
			}
		})
	}
}

// TestPrintBillingInfo_DoesNotPanic verifies that printBillingInfo doesn't panic with various inputs.
func TestPrintBillingInfo_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name string
		info *client.BillingInfo
	}{
		{
			name: "minimal info",
			info: &client.BillingInfo{
				Plan:         "free",
				Status:       "active",
				MonitorCount: 0,
				MonitorLimit: 10,
			},
		},
		{
			name: "full info",
			info: func() *client.BillingInfo {
				trialEnd := "2026-02-15T00:00:00Z"
				nextBilling := "2026-03-01T00:00:00Z"
				paymentMethod := "Visa ****4242"
				amount := 2900
				return &client.BillingInfo{
					Plan:           "team",
					Status:         "active",
					BillingEmail:   "billing@test.com",
					CurrentPeriod:  "monthly",
					MonitorCount:   100,
					MonitorLimit:   500,
					TrialEndsAt:    &trialEnd,
					NextBillingAt:  &nextBilling,
					PaymentMethod:  &paymentMethod,
					AmountCents:    &amount,
					Currency:       "USD",
					CancelAtPeriod: false,
				}
			}(),
		},
		{
			name: "trial subscription",
			info: func() *client.BillingInfo {
				trialEnd := "2026-02-15T00:00:00Z"
				return &client.BillingInfo{
					Plan:         "starter",
					Status:       "trialing",
					MonitorCount: 5,
					MonitorLimit: 25,
					TrialEndsAt:  &trialEnd,
				}
			}(),
		},
		{
			name: "canceled subscription",
			info: &client.BillingInfo{
				Plan:           "pro",
				Status:         "active",
				MonitorCount:   50,
				MonitorLimit:   100,
				CancelAtPeriod: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printBillingInfo panicked: %v", r)
				}
			}()
			printBillingInfo(tt.info)
		})
	}
}
