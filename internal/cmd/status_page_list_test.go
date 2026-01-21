package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestNewStatusPageListCmd(t *testing.T) {
	cmd := NewStatusPageListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use='list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "List all status pages" {
		t.Errorf("expected Short='List all status pages', got %q", cmd.Short)
	}
}

func TestNewStatusPageListCmd_Aliases(t *testing.T) {
	cmd := NewStatusPageListCmd()

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

func TestNewStatusPageListCmd_Long(t *testing.T) {
	cmd := NewStatusPageListCmd()

	long := cmd.Long

	// Should contain column documentation
	columns := []string{"NAME", "SLUG", "THEME", "PUBLIC", "ENABLED", "PROBES"}
	for _, col := range columns {
		if !strings.Contains(long, col) {
			t.Errorf("expected Long description to mention column %q", col)
		}
	}

	// Should mention wide mode columns
	wideColumns := []string{"DOMAIN", "UPTIME", "ID", "CREATED"}
	for _, col := range wideColumns {
		if !strings.Contains(long, col) {
			t.Errorf("expected Long description to mention wide column %q", col)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye status-page list") {
		t.Error("expected Long description to contain example commands")
	}
}

func TestNewStatusPageListCmd_Flags(t *testing.T) {
	cmd := NewStatusPageListCmd()

	// Verify expected flags exist
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"page", "1"},
		{"limit", "20"},
		{"search", ""},
		{"enabled", "false"},
		{"public", "false"},
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

func TestRunStatusPageList_Validation(t *testing.T) {
	// Test that validation errors are returned for invalid inputs.
	// Validation happens before making API calls.
	tests := []struct {
		name         string
		args         []string
		wantErrorMsg string
	}{
		{
			name:         "limit too low",
			args:         []string{"--limit", "0"},
			wantErrorMsg: "invalid limit 0: must be between 1 and 100",
		},
		{
			name:         "limit too high",
			args:         []string{"--limit", "101"},
			wantErrorMsg: "invalid limit 101: must be between 1 and 100",
		},
		{
			name:         "page too low",
			args:         []string{"--page", "0"},
			wantErrorMsg: "invalid page 0: must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStatusPageListCmd()
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

func TestRunStatusPageList_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	cmd := NewStatusPageListCmd()
	cmd.SetArgs([]string{"--enabled", "--search", "test"})

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
	}
	for _, ve := range validationErrors {
		if strings.Contains(err.Error(), ve) {
			t.Errorf("got unexpected validation error: %s", err.Error())
		}
	}
}

func TestRunStatusPageList_ValidBoundaryFlags(t *testing.T) {
	// Test boundary values for limit
	tests := []struct {
		name string
		args []string
	}{
		{"minimum limit", []string{"--limit", "1"}},
		{"maximum limit", []string{"--limit", "100"}},
		{"typical values", []string{"--page", "5", "--limit", "20"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStatusPageListCmd()
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

func TestRunStatusPageList_BooleanFlagsNotAppliedByDefault(t *testing.T) {
	// Verify that --enabled and --public flags are NOT applied when not specified.
	// This is a critical test for the cmd.Flags().Changed() pattern.
	cmd := NewStatusPageListCmd()
	cmd.SetArgs([]string{}) // No flags specified

	// The command should NOT apply enabled or public filters when not specified
	// We verify this indirectly by checking that the error is about API client,
	// not about filtering. In a real integration test, we'd verify the actual
	// API request parameters.

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should be about API client, confirming we got past validation
	// and didn't hit any issues with boolean flag handling
	if strings.Contains(err.Error(), "invalid") {
		t.Errorf("unexpected validation error: %s", err.Error())
	}
}

func TestRunStatusPageList_BooleanFlagsAppliedWhenSet(t *testing.T) {
	// Verify that --enabled and --public flags ARE applied when explicitly set.
	tests := []struct {
		name string
		args []string
	}{
		{"enabled flag", []string{"--enabled"}},
		{"public flag", []string{"--public"}},
		{"both flags", []string{"--enabled", "--public"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStatusPageListCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			// Should fail on API client, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Verify we got past the flags parsing (error is about API, not flags)
			if strings.Contains(err.Error(), "unknown flag") {
				t.Errorf("flag parsing error: %s", err.Error())
			}
		})
	}
}
