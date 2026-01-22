// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewIncidentDeleteCmd(t *testing.T) {
	cmd := NewIncidentDeleteCmd()

	if cmd.Use != "delete" {
		t.Errorf("expected Use='delete', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Delete an incident from a status page" {
		t.Errorf("expected Short='Delete an incident from a status page', got %q", cmd.Short)
	}
}

func TestNewIncidentDeleteCmd_Flags(t *testing.T) {
	cmd := NewIncidentDeleteCmd()

	// Verify expected flags exist with correct defaults
	flags := []struct {
		name         string
		shorthand    string
		defaultValue string
	}{
		{"status-page-id", "s", "0"},
		{"incident-id", "i", "0"},
		{"force", "f", "false"},
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
		if flag.Shorthand != f.shorthand {
			t.Errorf("flag %q: expected shorthand %q, got %q", f.name, f.shorthand, flag.Shorthand)
		}
	}
}

func TestNewIncidentDeleteCmd_Long(t *testing.T) {
	cmd := NewIncidentDeleteCmd()

	long := cmd.Long

	// Should contain required flags documentation
	requiredFlags := []string{"--status-page-id", "--incident-id"}
	for _, flag := range requiredFlags {
		if !strings.Contains(long, flag) {
			t.Errorf("expected Long description to mention flag %q", flag)
		}
	}

	// Should contain optional flags documentation
	optionalFlags := []string{"--force"}
	for _, flag := range optionalFlags {
		if !strings.Contains(long, flag) {
			t.Errorf("expected Long description to mention flag %q", flag)
		}
	}

	// Should contain warning about irreversible action
	if !strings.Contains(long, "WARNING") {
		t.Error("expected Long description to contain WARNING about irreversible action")
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye incident delete") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention difference between delete and resolve
	if !strings.Contains(long, "Resolve") {
		t.Error("expected Long description to mention difference between delete and resolve")
	}
}

func TestRunIncidentDelete_RequiredFlags(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewIncidentDeleteCmd()
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

func TestRunIncidentDelete_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	cmd := NewIncidentDeleteCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456"})

	err := cmd.Execute()

	// Should fail on API client initialization, not validation
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be a validation error about required flags
	if strings.Contains(err.Error(), "required flag") {
		t.Errorf("got unexpected validation error: %s", err.Error())
	}
}

func TestRunIncidentDelete_ForceFlag(t *testing.T) {
	// Test that force flag is accepted
	cmd := NewIncidentDeleteCmd()
	cmd.SetArgs([]string{
		"--status-page-id", "123",
		"--incident-id", "456",
		"--force",
	})

	err := cmd.Execute()

	// Should fail on API client, not flag parsing
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be about flag parsing
	if strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("got unexpected unknown flag error: %s", err.Error())
	}
}

func TestRunIncidentDelete_ShortFlags(t *testing.T) {
	// Test that short flags are accepted
	cmd := NewIncidentDeleteCmd()
	cmd.SetArgs([]string{"-s", "123", "-i", "456", "-f"})

	err := cmd.Execute()

	// Should fail on API client, not flag parsing
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be about flag parsing
	if strings.Contains(err.Error(), "unknown shorthand flag") {
		t.Errorf("got unexpected shorthand flag error: %s", err.Error())
	}
}

func TestNewIncidentDeleteCmd_RequiredFlagMarking(t *testing.T) {
	cmd := NewIncidentDeleteCmd()

	// Verify that status-page-id and incident-id are marked as required
	statusPageFlag := cmd.Flags().Lookup("status-page-id")
	incidentFlag := cmd.Flags().Lookup("incident-id")

	if statusPageFlag == nil {
		t.Fatal("status-page-id flag not found")
	}
	if incidentFlag == nil {
		t.Fatal("incident-id flag not found")
	}

	// Check by attempting to execute without required flags
	cmd2 := NewIncidentDeleteCmd()
	cmd2.SetArgs([]string{"--incident-id", "1"})
	err := cmd2.Execute()
	if err == nil || !strings.Contains(err.Error(), "status-page-id") {
		t.Error("status-page-id should be marked as required")
	}

	cmd3 := NewIncidentDeleteCmd()
	cmd3.SetArgs([]string{"--status-page-id", "1"})
	err = cmd3.Execute()
	if err == nil || !strings.Contains(err.Error(), "incident-id") {
		t.Error("incident-id should be marked as required")
	}
}

func TestNewIncidentDeleteCmd_ForceFlagOptional(t *testing.T) {
	// Test that command works without force flag (will fail on API, not validation)
	cmd := NewIncidentDeleteCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456"})

	err := cmd.Execute()

	// Should fail on API client, not because force is missing
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT mention force being required
	if strings.Contains(err.Error(), "force") && strings.Contains(err.Error(), "required") {
		t.Errorf("force should be optional, got error: %s", err.Error())
	}
}
