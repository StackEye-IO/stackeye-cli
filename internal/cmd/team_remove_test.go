// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewTeamRemoveCmd verifies that the team remove command is properly constructed.
func TestNewTeamRemoveCmd(t *testing.T) {
	cmd := NewTeamRemoveCmd()

	if cmd.Use != "remove" {
		t.Errorf("expected Use to be 'remove', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "Remove a team member") {
		t.Error("expected Long description to mention 'Remove a team member'")
	}
}

// TestNewTeamRemoveCmd_HelpContainsWarning verifies that the help text warns about restrictions.
func TestNewTeamRemoveCmd_HelpContainsWarning(t *testing.T) {
	cmd := NewTeamRemoveCmd()

	// Verify restriction is documented
	if !strings.Contains(cmd.Long, "cannot remove") {
		t.Error("expected Long description to mention restrictions")
	}
}

// TestNewTeamRemoveCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewTeamRemoveCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewTeamRemoveCmd()

	// Verify examples are present
	examples := []string{
		"stackeye team remove",
		"--member-id",
		"--email",
		"--force",
		"-o json",
	}
	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

// TestNewTeamRemoveCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewTeamRemoveCmd_RunEIsSet(t *testing.T) {
	cmd := NewTeamRemoveCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewTeamRemoveCmd_HasFlags verifies flags are defined.
func TestNewTeamRemoveCmd_HasFlags(t *testing.T) {
	cmd := NewTeamRemoveCmd()

	// Check --member-id flag
	memberIDFlag := cmd.Flags().Lookup("member-id")
	if memberIDFlag == nil {
		t.Error("expected --member-id flag to be defined")
	}

	// Check --email flag
	emailFlag := cmd.Flags().Lookup("email")
	if emailFlag == nil {
		t.Error("expected --email flag to be defined")
	}

	// Check --force flag
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("expected --force flag to be defined")
	}
}

// TestValidateTeamRemoveFlags_ValidValues verifies valid flag values are accepted.
func TestValidateTeamRemoveFlags_ValidValues(t *testing.T) {
	tests := []struct {
		name  string
		flags *teamRemoveFlags
	}{
		{
			name:  "valid member-id only",
			flags: &teamRemoveFlags{memberID: 42},
		},
		{
			name:  "valid member-id with force",
			flags: &teamRemoveFlags{memberID: 1, force: true},
		},
		{
			name:  "valid email only",
			flags: &teamRemoveFlags{email: "user@company.io"},
		},
		{
			name:  "valid email with force",
			flags: &teamRemoveFlags{email: "user@company.io", force: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamRemoveFlags(tt.flags)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestValidateTeamRemoveFlags_InvalidValues verifies invalid flag values are rejected.
func TestValidateTeamRemoveFlags_InvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		flags       *teamRemoveFlags
		errContains string
	}{
		{
			name:        "no member identification",
			flags:       &teamRemoveFlags{memberID: 0, email: ""},
			errContains: "either --member-id or --email is required",
		},
		{
			name:        "both member-id and email",
			flags:       &teamRemoveFlags{memberID: 42, email: "user@company.io"},
			errContains: "specify either --member-id or --email, not both",
		},
		{
			name:        "both with force",
			flags:       &teamRemoveFlags{memberID: 42, email: "user@company.io", force: true},
			errContains: "specify either --member-id or --email, not both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamRemoveFlags(tt.flags)
			if err == nil {
				t.Error("expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

// TestTeamRemoveTimeout verifies the timeout constant is reasonable.
func TestTeamRemoveTimeout(t *testing.T) {
	// Timeout should be at least 10 seconds for network operations
	if teamRemoveTimeout.Seconds() < 10 {
		t.Errorf("expected timeout to be at least 10 seconds, got %v", teamRemoveTimeout)
	}

	// Timeout should not exceed 60 seconds for a simple operation
	if teamRemoveTimeout.Seconds() > 60 {
		t.Errorf("expected timeout to be at most 60 seconds, got %v", teamRemoveTimeout)
	}
}
