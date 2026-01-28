// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewBillingStatusCmd verifies that the billing status command is properly constructed.
func TestNewBillingStatusCmd(t *testing.T) {
	cmd := NewBillingStatusCmd()

	if cmd.Use != "status" {
		t.Errorf("expected Use='status', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Show current billing status" {
		t.Errorf("expected Short='Show current billing status', got %q", cmd.Short)
	}
}

// TestNewBillingStatusCmd_Long verifies the Long description contains key information.
func TestNewBillingStatusCmd_Long(t *testing.T) {
	cmd := NewBillingStatusCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"lightweight",
		"Subscription status",
		"Stripe",
	}
	for _, feature := range features {
		if !strings.Contains(long, feature) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye billing status") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention output formats
	if !strings.Contains(long, "json") {
		t.Error("expected Long description to mention JSON output option")
	}
}

// TestNewBillingStatusCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewBillingStatusCmd_RunEIsSet(t *testing.T) {
	cmd := NewBillingStatusCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewBillingStatusCmd_NoAliases verifies that status command has no aliases.
// The 'status' command should be the dedicated command, not an alias.
func TestNewBillingStatusCmd_NoAliases(t *testing.T) {
	cmd := NewBillingStatusCmd()

	if len(cmd.Aliases) != 0 {
		t.Errorf("expected no aliases, got %d: %v", len(cmd.Aliases), cmd.Aliases)
	}
}

// TestFormatBillingStatusValue verifies subscription status formatting.
func TestFormatBillingStatusValue(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"active", "active", "[*] Active"},
		{"trialing", "trialing", "[~] Trial"},
		{"past due", "past_due", "[!] Past Due"},
		{"canceled lowercase", "canceled", "[ ] Canceled"},
		{"cancelled UK spelling", "cancelled", "[ ] Canceled"},
		{"incomplete", "incomplete", "[.] Incomplete"},
		{"incomplete_expired", "incomplete_expired", "[ ] Expired"},
		{"unpaid", "unpaid", "[!] Unpaid"},
		{"paused", "paused", "[-] Paused"},
		{"none", "none", "[ ] None"},
		{"empty string", "", "[ ] None"},
		{"unknown status", "pending", "pending"},
		{"uppercase active", "ACTIVE", "[*] Active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBillingStatusValue(tt.status)
			if result != tt.expected {
				t.Errorf("formatBillingStatusValue(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

// TestFormatStripeStatus verifies Stripe status formatting.
func TestFormatStripeStatus(t *testing.T) {
	tests := []struct {
		name        string
		hasCustomer bool
		expected    string
	}{
		{"has customer", true, "Connected"},
		{"no customer", false, "Not configured"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStripeStatus(tt.hasCustomer)
			if result != tt.expected {
				t.Errorf("formatStripeStatus(%v) = %q, want %q", tt.hasCustomer, result, tt.expected)
			}
		})
	}
}

// TestPrintBillingStatus_DoesNotPanic verifies that printBillingStatus doesn't panic.
func TestPrintBillingStatus_DoesNotPanic(t *testing.T) {
	tests := []struct {
		name   string
		status *client.BillingStatus
	}{
		{
			name: "active subscription",
			status: &client.BillingStatus{
				HasStripeCustomer:  true,
				SubscriptionStatus: "active",
			},
		},
		{
			name: "trial subscription",
			status: &client.BillingStatus{
				HasStripeCustomer:  true,
				SubscriptionStatus: "trialing",
			},
		},
		{
			name: "no subscription",
			status: &client.BillingStatus{
				HasStripeCustomer:  false,
				SubscriptionStatus: "none",
			},
		},
		{
			name: "past due",
			status: &client.BillingStatus{
				HasStripeCustomer:  true,
				SubscriptionStatus: "past_due",
			},
		},
		{
			name: "canceled",
			status: &client.BillingStatus{
				HasStripeCustomer:  true,
				SubscriptionStatus: "canceled",
			},
		},
		{
			name: "empty status",
			status: &client.BillingStatus{
				HasStripeCustomer:  false,
				SubscriptionStatus: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printBillingStatus panicked: %v", r)
				}
			}()
			printBillingStatus(tt.status)
		})
	}
}
