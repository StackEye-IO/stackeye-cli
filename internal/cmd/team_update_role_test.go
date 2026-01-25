// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewTeamUpdateRoleCmd verifies that the team update-role command is properly constructed.
func TestNewTeamUpdateRoleCmd(t *testing.T) {
	cmd := NewTeamUpdateRoleCmd()

	if cmd.Use != "update-role" {
		t.Errorf("expected Use to be 'update-role', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "Update the role") {
		t.Error("expected Long description to mention 'Update the role'")
	}
}

// TestNewTeamUpdateRoleCmd_HelpContainsRoles verifies that the help text explains user roles.
func TestNewTeamUpdateRoleCmd_HelpContainsRoles(t *testing.T) {
	cmd := NewTeamUpdateRoleCmd()

	// Verify all roles are documented
	roles := []string{"owner", "admin", "member", "viewer"}
	for _, role := range roles {
		if !strings.Contains(cmd.Long, role) {
			t.Errorf("expected Long description to document role %q", role)
		}
	}
}

// TestNewTeamUpdateRoleCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewTeamUpdateRoleCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewTeamUpdateRoleCmd()

	// Verify examples are present
	examples := []string{
		"stackeye team update-role",
		"--member-id",
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

// TestNewTeamUpdateRoleCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewTeamUpdateRoleCmd_RunEIsSet(t *testing.T) {
	cmd := NewTeamUpdateRoleCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewTeamUpdateRoleCmd_HasFlags verifies flags are defined.
func TestNewTeamUpdateRoleCmd_HasFlags(t *testing.T) {
	cmd := NewTeamUpdateRoleCmd()

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

	// Check --role flag
	roleFlag := cmd.Flags().Lookup("role")
	if roleFlag == nil {
		t.Error("expected --role flag to be defined")
	}
}

// TestValidateTeamUpdateRoleFlags_ValidValues verifies valid flag values are accepted.
func TestValidateTeamUpdateRoleFlags_ValidValues(t *testing.T) {
	tests := []struct {
		name  string
		flags *teamUpdateRoleFlags
	}{
		{
			name:  "valid member-id and admin role",
			flags: &teamUpdateRoleFlags{memberID: 42, role: "admin"},
		},
		{
			name:  "valid member-id and member role",
			flags: &teamUpdateRoleFlags{memberID: 1, role: "member"},
		},
		{
			name:  "valid member-id and viewer role",
			flags: &teamUpdateRoleFlags{memberID: 100, role: "viewer"},
		},
		{
			name:  "valid member-id and owner role",
			flags: &teamUpdateRoleFlags{memberID: 1, role: "owner"},
		},
		{
			name:  "valid email and admin role",
			flags: &teamUpdateRoleFlags{email: "user@company.io", role: "admin"},
		},
		{
			name:  "valid email and member role",
			flags: &teamUpdateRoleFlags{email: "user@company.io", role: "member"},
		},
		{
			name:  "uppercase role with member-id",
			flags: &teamUpdateRoleFlags{memberID: 42, role: "ADMIN"},
		},
		{
			name:  "mixed case role with email",
			flags: &teamUpdateRoleFlags{email: "user@company.io", role: "Member"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamUpdateRoleFlags(tt.flags)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestValidateTeamUpdateRoleFlags_InvalidValues verifies invalid flag values are rejected.
func TestValidateTeamUpdateRoleFlags_InvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		flags       *teamUpdateRoleFlags
		errContains string
	}{
		{
			name:        "no member identification",
			flags:       &teamUpdateRoleFlags{memberID: 0, email: "", role: "admin"},
			errContains: "either --member-id or --email is required",
		},
		{
			name:        "both member-id and email",
			flags:       &teamUpdateRoleFlags{memberID: 42, email: "user@company.io", role: "admin"},
			errContains: "specify either --member-id or --email, not both",
		},
		{
			name:        "empty role with member-id",
			flags:       &teamUpdateRoleFlags{memberID: 42, role: ""},
			errContains: "role is required",
		},
		{
			name:        "empty role with email",
			flags:       &teamUpdateRoleFlags{email: "user@company.io", role: ""},
			errContains: "role is required",
		},
		{
			name:        "invalid role with member-id",
			flags:       &teamUpdateRoleFlags{memberID: 42, role: "superadmin"},
			errContains: "invalid role",
		},
		{
			name:        "invalid role typo",
			flags:       &teamUpdateRoleFlags{memberID: 42, role: "adnin"},
			errContains: "invalid role",
		},
		{
			name:        "invalid role with email",
			flags:       &teamUpdateRoleFlags{email: "user@company.io", role: "moderator"},
			errContains: "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamUpdateRoleFlags(tt.flags)
			if err == nil {
				t.Error("expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

// TestTeamUpdateRoleTimeout verifies the timeout constant is reasonable.
func TestTeamUpdateRoleTimeout(t *testing.T) {
	// Timeout should be at least 10 seconds for network operations
	if teamUpdateRoleTimeout.Seconds() < 10 {
		t.Errorf("expected timeout to be at least 10 seconds, got %v", teamUpdateRoleTimeout)
	}

	// Timeout should not exceed 60 seconds for a simple operation
	if teamUpdateRoleTimeout.Seconds() > 60 {
		t.Errorf("expected timeout to be at most 60 seconds, got %v", teamUpdateRoleTimeout)
	}
}
