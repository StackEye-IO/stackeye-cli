package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestNewIncidentListCmd(t *testing.T) {
	cmd := NewIncidentListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "List incidents for a status page" {
		t.Errorf("expected Short='List incidents for a status page', got %q", cmd.Short)
	}
}

func TestNewIncidentListCmd_Aliases(t *testing.T) {
	cmd := NewIncidentListCmd()

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

func TestNewIncidentListCmd_Long(t *testing.T) {
	cmd := NewIncidentListCmd()

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
	if !strings.Contains(long, "stackeye incident list") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention valid status values
	statusValues := []string{"investigating", "identified", "monitoring", "resolved"}
	for _, status := range statusValues {
		if !strings.Contains(long, status) {
			t.Errorf("expected Long description to mention status %q", status)
		}
	}
}

func TestNewIncidentListCmd_Flags(t *testing.T) {
	cmd := NewIncidentListCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status-page-id", "0"},
		{"page", "1"},
		{"limit", "20"},
		{"status", ""},
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

func TestRunIncidentList_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Validation happens before making API calls.
	tests := []struct {
		name         string
		args         []string
		wantErrorMsg string
	}{
		{
			name:         "status-page-id required",
			args:         []string{},
			wantErrorMsg: "required flag(s) \"status-page-id\" not set",
		},
		{
			name:         "limit too low",
			args:         []string{"--status-page-id", "1", "--limit", "0"},
			wantErrorMsg: "invalid limit 0: must be between 1 and 100",
		},
		{
			name:         "limit too high",
			args:         []string{"--status-page-id", "1", "--limit", "101"},
			wantErrorMsg: "invalid limit 101: must be between 1 and 100",
		},
		{
			name:         "page too low",
			args:         []string{"--status-page-id", "1", "--page", "0"},
			wantErrorMsg: "invalid page 0: must be at least 1",
		},
		{
			name:         "invalid status value",
			args:         []string{"--status-page-id", "1", "--status", "invalid"},
			wantErrorMsg: "for --status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentListCmd()
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

func TestRunIncidentList_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	cmd := NewIncidentListCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--status", "investigating"})

	err := cmd.Execute()

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
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}

func TestRunIncidentList_ValidBoundaryFlags(t *testing.T) {
	// Test boundary values for limit
	tests := []struct {
		name string
		args []string
	}{
		{"minimum limit", []string{"--status-page-id", "1", "--limit", "1"}},
		{"maximum limit", []string{"--status-page-id", "1", "--limit", "100"}},
		{"typical values", []string{"--status-page-id", "1", "--page", "5", "--limit", "20"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentListCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			// Should fail on API client, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Error should not be validation related
			if strings.Contains(err.Error(), "invalid limit") || strings.Contains(err.Error(), "invalid page") {
				t.Errorf("got unexpected validation error for valid input: %s", err.Error())
			}
		})
	}
}

func TestRunIncidentList_ValidStatusValues(t *testing.T) {
	// Test all valid status values
	validStatuses := []string{"investigating", "identified", "monitoring", "resolved"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			cmd := NewIncidentListCmd()
			cmd.SetArgs([]string{"--status-page-id", "1", "--status", status})

			err := cmd.Execute()

			// Should fail on API client, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Error should not be about invalid status
			if strings.Contains(err.Error(), "invalid status") {
				t.Errorf("status %q should be valid, got error: %s", status, err.Error())
			}
		})
	}
}

func TestRunIncidentList_StatusFilterIsOptional(t *testing.T) {
	// Verify that --status filter is optional
	cmd := NewIncidentListCmd()
	cmd.SetArgs([]string{"--status-page-id", "123"}) // No --status flag

	err := cmd.Execute()

	// Should fail on API client, not missing flag
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should be about API client, not missing status flag
	if strings.Contains(err.Error(), "status") && strings.Contains(err.Error(), "required") {
		t.Errorf("--status should be optional, got error: %s", err.Error())
	}
}
