package cmd

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestNewAlertListCmd(t *testing.T) {
	cmd := NewAlertListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "List all monitoring alerts" {
		t.Errorf("expected Short='List all monitoring alerts', got %q", cmd.Short)
	}
}

func TestNewAlertListCmd_Aliases(t *testing.T) {
	cmd := NewAlertListCmd()

	expectedAliases := []string{"ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewAlertListCmd_Long(t *testing.T) {
	cmd := NewAlertListCmd()

	long := cmd.Long

	// Should contain status documentation
	statuses := []string{"active", "acknowledged", "resolved"}
	for _, status := range statuses {
		if !strings.Contains(long, status) {
			t.Errorf("expected Long description to mention status %q", status)
		}
	}

	// Should contain severity documentation
	severities := []string{"critical", "warning", "info"}
	for _, severity := range severities {
		if !strings.Contains(long, severity) {
			t.Errorf("expected Long description to mention severity %q", severity)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye alert list") {
		t.Error("expected Long description to contain example commands")
	}
}

func TestNewAlertListCmd_Flags(t *testing.T) {
	cmd := NewAlertListCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"status", "", ""},
		{"severity", "", ""},
		{"probe", "", ""},
		{"page", "", "1"},
		{"limit", "", "20"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected flag %q to exist", f.name)
			continue
		}
		if flag.DefValue != f.defaultValue {
			t.Errorf("flag %q: expected default %q, got %q", f.name, f.defaultValue, flag.DefValue)
		}
	}
}

func TestRunAlertList_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Since runAlertList requires an API client, validation happens first
	// and will fail before making API calls for invalid inputs.
	tests := []struct {
		name         string
		limit        int
		page         int
		status       string
		severity     string
		probeID      string
		wantErrorMsg string
	}{
		{
			name:         "limit too low",
			limit:        0,
			page:         1,
			wantErrorMsg: "invalid limit 0: must be between 1 and 100",
		},
		{
			name:         "limit too high",
			limit:        101,
			page:         1,
			wantErrorMsg: "invalid limit 101: must be between 1 and 100",
		},
		{
			name:         "page too low",
			limit:        20,
			page:         0,
			wantErrorMsg: "invalid page 0: must be at least 1",
		},
		{
			name:         "invalid status",
			limit:        20,
			page:         1,
			status:       "badstatus",
			wantErrorMsg: `invalid value "badstatus" for --status: must be one of: active, acknowledged, resolved`,
		},
		{
			name:         "invalid severity",
			limit:        20,
			page:         1,
			severity:     "badseverity",
			wantErrorMsg: `invalid value "badseverity" for --severity: must be one of: critical, warning, info`,
		},
		{
			name:         "invalid probe ID",
			limit:        20,
			page:         1,
			probeID:      "not-a-uuid",
			wantErrorMsg: `invalid probe ID "not-a-uuid"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &alertListFlags{
				page:     tt.page,
				limit:    tt.limit,
				status:   tt.status,
				severity: tt.severity,
				probeID:  tt.probeID,
			}

			// Call runAlertList with a background context.
			// It should fail on validation before needing API client.
			err := runAlertList(context.Background(), flags)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

func TestRunAlertList_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	flags := &alertListFlags{
		page:     1,
		limit:    20,
		status:   "active",
		severity: "critical",
	}

	err := runAlertList(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error
	validationErrors := []string{
		"invalid limit",
		"invalid page",
		"invalid status",
		"invalid severity",
		"invalid probe ID",
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}

func TestRunAlertList_ValidProbeID(t *testing.T) {
	// Test that a valid UUID probe ID passes validation
	flags := &alertListFlags{
		page:    1,
		limit:   20,
		probeID: "123e4567-e89b-12d3-a456-426614174000", // Valid UUID
	}

	err := runAlertList(context.Background(), flags)

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a probe ID validation error
	if strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("got unexpected validation error for valid probe ID: %s", err.Error())
	}
}
