package cmd

import (
	"strings"
	"testing"
)

func TestNewIncidentGetCmd(t *testing.T) {
	cmd := NewIncidentGetCmd()

	if cmd.Use != "get" {
		t.Errorf("expected Use='get', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Get details of a specific incident" {
		t.Errorf("expected Short='Get details of a specific incident', got %q", cmd.Short)
	}
}

func TestNewIncidentGetCmd_Long(t *testing.T) {
	cmd := NewIncidentGetCmd()

	long := cmd.Long

	// Should contain column documentation
	columns := []string{"ID", "TITLE", "STATUS", "IMPACT", "CREATED"}
	for _, col := range columns {
		if !strings.Contains(long, col) {
			t.Errorf("expected Long description to mention column %q", col)
		}
	}

	// Should mention wide mode columns
	wideColumns := []string{"UPDATED", "RESOLVED"}
	for _, col := range wideColumns {
		if !strings.Contains(long, col) {
			t.Errorf("expected Long description to mention wide column %q", col)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye incident get") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention valid status values
	statusValues := []string{"investigating", "identified", "monitoring", "resolved"}
	for _, status := range statusValues {
		if !strings.Contains(long, status) {
			t.Errorf("expected Long description to mention status %q", status)
		}
	}

	// Should mention impact levels
	impactLevels := []string{"none", "minor", "major", "critical"}
	for _, impact := range impactLevels {
		if !strings.Contains(long, impact) {
			t.Errorf("expected Long description to mention impact %q", impact)
		}
	}
}

func TestNewIncidentGetCmd_Flags(t *testing.T) {
	cmd := NewIncidentGetCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status-page-id", "0"},
		{"incident-id", "0"},
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

func TestRunIncidentGet_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Validation happens before making API calls.
	tests := []struct {
		name         string
		args         []string
		wantErrorMsg string
	}{
		{
			name:         "status-page-id required",
			args:         []string{"--incident-id", "456"},
			wantErrorMsg: "required flag(s) \"status-page-id\" not set",
		},
		{
			name:         "incident-id required",
			args:         []string{"--status-page-id", "123"},
			wantErrorMsg: "required flag(s) \"incident-id\" not set",
		},
		{
			name:         "both flags required",
			args:         []string{},
			wantErrorMsg: "required flag(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentGetCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

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

func TestRunIncidentGet_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	cmd := NewIncidentGetCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456"})

	err := cmd.Execute()

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error about required flags
	if strings.Contains(err.Error(), "required flag") {
		t.Errorf("got unexpected required flag error: %s", err.Error())
	}
}

func TestRunIncidentGet_ZeroValues(t *testing.T) {
	// Test that zero values are rejected even if flags are technically "set"
	tests := []struct {
		name         string
		args         []string
		wantErrorMsg string
	}{
		{
			name:         "status-page-id zero",
			args:         []string{"--status-page-id", "0", "--incident-id", "456"},
			wantErrorMsg: "--status-page-id is required",
		},
		{
			name:         "incident-id zero",
			args:         []string{"--status-page-id", "123", "--incident-id", "0"},
			wantErrorMsg: "--incident-id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentGetCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

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

func TestRunIncidentGet_LargeIDs(t *testing.T) {
	// Test that large valid IDs pass validation (will fail later on API client)
	cmd := NewIncidentGetCmd()
	cmd.SetArgs([]string{"--status-page-id", "999999999", "--incident-id", "888888888"})

	err := cmd.Execute()

	// Should fail on API client, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should not be about invalid IDs
	if strings.Contains(err.Error(), "is required") {
		t.Errorf("got unexpected validation error for valid input: %s", err.Error())
	}
}
