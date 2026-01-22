// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

func TestNewIncidentResolveCmd(t *testing.T) {
	cmd := NewIncidentResolveCmd()

	if cmd.Use != "resolve" {
		t.Errorf("expected Use='resolve', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Resolve an incident on a status page" {
		t.Errorf("expected Short='Resolve an incident on a status page', got %q", cmd.Short)
	}
}

func TestNewIncidentResolveCmd_Flags(t *testing.T) {
	cmd := NewIncidentResolveCmd()

	// Verify expected flags exist with correct defaults
	flags := []struct {
		name         string
		defaultValue string
	}{
		{"status-page-id", "0"},
		{"incident-id", "0"},
		{"message", ""},
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

func TestNewIncidentResolveCmd_Long(t *testing.T) {
	cmd := NewIncidentResolveCmd()

	long := cmd.Long

	// Should contain required flags documentation
	requiredFlags := []string{"--status-page-id", "--incident-id"}
	for _, flag := range requiredFlags {
		if !strings.Contains(long, flag) {
			t.Errorf("expected Long description to mention flag %q", flag)
		}
	}

	// Should contain optional flags documentation
	optionalFlags := []string{"--message"}
	for _, flag := range optionalFlags {
		if !strings.Contains(long, flag) {
			t.Errorf("expected Long description to mention flag %q", flag)
		}
	}

	// Should contain status workflow
	if !strings.Contains(long, "resolved") {
		t.Error("expected Long description to mention 'resolved' status")
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye incident resolve") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention what happens when resolving
	if !strings.Contains(long, "Resolution timestamp") {
		t.Error("expected Long description to mention resolution timestamp")
	}
}

func TestRunIncidentResolve_RequiredFlags(t *testing.T) {
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
			cmd := NewIncidentResolveCmd()
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

func TestRunIncidentResolve_ValidFlags(t *testing.T) {
	// Test that valid flags pass validation (will fail later on API client)
	cmd := NewIncidentResolveCmd()
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

func TestRunIncidentResolve_WithMessage(t *testing.T) {
	// Test that message flag is accepted
	cmd := NewIncidentResolveCmd()
	cmd.SetArgs([]string{
		"--status-page-id", "123",
		"--incident-id", "456",
		"--message", "Issue has been fixed by restarting the service.",
	})

	err := cmd.Execute()

	// Should fail on API client, not flag parsing
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT be about flag parsing
	if strings.Contains(err.Error(), "required flag") {
		t.Errorf("got unexpected flag parsing error: %s", err.Error())
	}
	if strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("got unexpected unknown flag error: %s", err.Error())
	}
}

func TestNewIncidentResolveCmd_RequiredFlagMarking(t *testing.T) {
	cmd := NewIncidentResolveCmd()

	// Verify that status-page-id and incident-id are marked as required
	statusPageFlag := cmd.Flags().Lookup("status-page-id")
	incidentFlag := cmd.Flags().Lookup("incident-id")

	if statusPageFlag == nil {
		t.Fatal("status-page-id flag not found")
	}
	if incidentFlag == nil {
		t.Fatal("incident-id flag not found")
	}

	// Check annotations for required marking
	// Cobra marks required flags with annotations
	if ann, ok := statusPageFlag.Annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok || len(ann) == 0 {
		// Alternative check: try to execute without the flag
		cmd2 := NewIncidentResolveCmd()
		cmd2.SetArgs([]string{"--incident-id", "1"})
		err := cmd2.Execute()
		if err == nil || !strings.Contains(err.Error(), "status-page-id") {
			t.Error("status-page-id should be marked as required")
		}
	}
}

func TestNewIncidentResolveCmd_MessageFlagOptional(t *testing.T) {
	// Test that command works without message flag (will fail on API, not validation)
	cmd := NewIncidentResolveCmd()
	cmd.SetArgs([]string{"--status-page-id", "123", "--incident-id", "456"})

	err := cmd.Execute()

	// Should fail on API client, not because message is missing
	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should NOT mention message being required
	if strings.Contains(err.Error(), "message") && strings.Contains(err.Error(), "required") {
		t.Errorf("message should be optional, got error: %s", err.Error())
	}
}
