// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewBillingPortalCmd verifies that the billing portal command is properly constructed.
func TestNewBillingPortalCmd(t *testing.T) {
	cmd := NewBillingPortalCmd()

	if cmd.Use != "portal" {
		t.Errorf("expected Use='portal', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Open Stripe billing portal" {
		t.Errorf("expected Short='Open Stripe billing portal', got %q", cmd.Short)
	}
}

// TestNewBillingPortalCmd_Long verifies the Long description contains key information.
func TestNewBillingPortalCmd_Long(t *testing.T) {
	cmd := NewBillingPortalCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"customer portal",
		"payment methods",
		"invoices",
		"subscription",
	}
	for _, feature := range features {
		if !strings.Contains(long, feature) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye billing portal") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention url-only flag
	if !strings.Contains(long, "--url-only") {
		t.Error("expected Long description to mention --url-only flag")
	}
}

// TestNewBillingPortalCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewBillingPortalCmd_RunEIsSet(t *testing.T) {
	cmd := NewBillingPortalCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewBillingPortalCmd_HasUrlOnlyFlag verifies the --url-only flag exists.
func TestNewBillingPortalCmd_HasUrlOnlyFlag(t *testing.T) {
	cmd := NewBillingPortalCmd()

	flag := cmd.Flags().Lookup("url-only")
	if flag == nil {
		t.Error("expected --url-only flag to be defined")
		return
	}

	if flag.DefValue != "false" {
		t.Errorf("expected --url-only default value to be 'false', got %q", flag.DefValue)
	}
}

// TestBrowserOpenerVariable verifies that browserOpener is set to a valid function.
func TestBrowserOpenerVariable(t *testing.T) {
	if browserOpener == nil {
		t.Error("expected browserOpener to be set")
	}
}

// TestNewBillingPortalCmd_NoAliases verifies that portal command has no aliases.
func TestNewBillingPortalCmd_NoAliases(t *testing.T) {
	cmd := NewBillingPortalCmd()

	if len(cmd.Aliases) != 0 {
		t.Errorf("expected no aliases, got %d: %v", len(cmd.Aliases), cmd.Aliases)
	}
}

// TestBrowserOpenerOverride verifies that browserOpener can be overridden for testing.
func TestBrowserOpenerOverride(t *testing.T) {
	// Save original
	original := browserOpener
	defer func() { browserOpener = original }()

	// Override with test function
	called := false
	calledURL := ""
	browserOpener = func(url string) error {
		called = true
		calledURL = url
		return nil
	}

	// Call the override
	testURL := "https://example.com"
	err := browserOpener(testURL)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected browserOpener override to be called")
	}
	if calledURL != testURL {
		t.Errorf("expected browserOpener to be called with %q, got %q", testURL, calledURL)
	}
}

// TestPortalFlagsStruct verifies the portalFlags struct exists and has expected fields.
func TestPortalFlagsStruct(t *testing.T) {
	// Save original flags
	originalFlags := portalFlags
	defer func() { portalFlags = originalFlags }()

	// Verify we can access the urlOnly field
	portalFlags.urlOnly = true
	if !portalFlags.urlOnly {
		t.Error("expected portalFlags.urlOnly to be settable to true")
	}
	portalFlags.urlOnly = false
	if portalFlags.urlOnly {
		t.Error("expected portalFlags.urlOnly to be settable to false")
	}
}

// TestBillingPortalTimeout verifies the timeout constant is reasonable.
func TestBillingPortalTimeout(t *testing.T) {
	// Timeout should be at least 10 seconds
	if billingPortalTimeout.Seconds() < 10 {
		t.Errorf("billingPortalTimeout too short: %v", billingPortalTimeout)
	}
	// Timeout should be at most 60 seconds
	if billingPortalTimeout.Seconds() > 60 {
		t.Errorf("billingPortalTimeout too long: %v", billingPortalTimeout)
	}
}
