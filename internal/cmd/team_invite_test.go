// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewTeamInviteCmd verifies that the team invite command is properly constructed.
func TestNewTeamInviteCmd(t *testing.T) {
	cmd := NewTeamInviteCmd()

	if cmd.Use != "invite" {
		t.Errorf("expected Use to be 'invite', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "team member") {
		t.Error("expected Long description to mention 'team member'")
	}
}

// TestNewTeamInviteCmd_HelpContainsRoles verifies that the help text explains user roles.
func TestNewTeamInviteCmd_HelpContainsRoles(t *testing.T) {
	cmd := NewTeamInviteCmd()

	// Verify all roles are documented
	roles := []string{"owner", "admin", "member", "viewer"}
	for _, role := range roles {
		if !strings.Contains(cmd.Long, role) {
			t.Errorf("expected Long description to document role %q", role)
		}
	}
}

// TestNewTeamInviteCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewTeamInviteCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewTeamInviteCmd()

	// Verify examples are present
	examples := []string{
		"stackeye team invite",
		"--email",
		"--role",
		"-o json",
	}
	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

// TestNewTeamInviteCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewTeamInviteCmd_RunEIsSet(t *testing.T) {
	cmd := NewTeamInviteCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewTeamInviteCmd_HasRequiredFlags verifies required flags are defined.
func TestNewTeamInviteCmd_HasRequiredFlags(t *testing.T) {
	cmd := NewTeamInviteCmd()

	// Check --email flag
	emailFlag := cmd.Flags().Lookup("email")
	if emailFlag == nil {
		t.Error("expected --email flag to be defined")
	}

	// Check --role flag
	roleFlag := cmd.Flags().Lookup("role")
	if roleFlag == nil {
		t.Error("expected --role flag to be defined")
	}
}

// TestValidateTeamInviteFlags_ValidValues verifies valid flag values are accepted.
func TestValidateTeamInviteFlags_ValidValues(t *testing.T) {
	tests := []struct {
		name  string
		flags *teamInviteFlags
	}{
		{
			name:  "valid admin invite",
			flags: &teamInviteFlags{email: "user@company.io", role: "admin"},
		},
		{
			name:  "valid member invite",
			flags: &teamInviteFlags{email: "user@company.io", role: "member"},
		},
		{
			name:  "valid viewer invite",
			flags: &teamInviteFlags{email: "user@company.io", role: "viewer"},
		},
		{
			name:  "valid owner invite",
			flags: &teamInviteFlags{email: "user@company.io", role: "owner"},
		},
		{
			name:  "uppercase role",
			flags: &teamInviteFlags{email: "user@company.io", role: "ADMIN"},
		},
		{
			name:  "mixed case role",
			flags: &teamInviteFlags{email: "user@company.io", role: "Member"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamInviteFlags(tt.flags)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestValidateTeamInviteFlags_InvalidValues verifies invalid flag values are rejected.
func TestValidateTeamInviteFlags_InvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		flags       *teamInviteFlags
		errContains string
	}{
		{
			name:        "empty email",
			flags:       &teamInviteFlags{email: "", role: "admin"},
			errContains: "email is required",
		},
		{
			name:        "invalid email no at",
			flags:       &teamInviteFlags{email: "usercompany.io", role: "admin"},
			errContains: "invalid email format",
		},
		{
			name:        "invalid email no dot",
			flags:       &teamInviteFlags{email: "user@companyio", role: "admin"},
			errContains: "invalid email format",
		},
		{
			name:        "empty role",
			flags:       &teamInviteFlags{email: "user@company.io", role: ""},
			errContains: "role is required",
		},
		{
			name:        "invalid role",
			flags:       &teamInviteFlags{email: "user@company.io", role: "superadmin"},
			errContains: "for --role",
		},
		{
			name:        "invalid role typo",
			flags:       &teamInviteFlags{email: "user@company.io", role: "adnin"},
			errContains: "for --role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamInviteFlags(tt.flags)
			if err == nil {
				t.Error("expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

// TestValidTeamRoles verifies the validTeamRoles slice contains expected values.
func TestValidTeamRoles(t *testing.T) {
	expected := []string{"owner", "admin", "member", "viewer"}

	if len(validTeamRoles) != len(expected) {
		t.Errorf("expected %d roles, got %d", len(expected), len(validTeamRoles))
	}

	for i, role := range expected {
		if validTeamRoles[i] != role {
			t.Errorf("expected role[%d] to be %q, got %q", i, role, validTeamRoles[i])
		}
	}
}
