// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewTeamRevokeInvitationCmd verifies that the team revoke-invitation command is properly constructed.
func TestNewTeamRevokeInvitationCmd(t *testing.T) {
	cmd := NewTeamRevokeInvitationCmd()

	if cmd.Use != "revoke-invitation" {
		t.Errorf("expected Use to be 'revoke-invitation', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "Revoke a pending invitation") {
		t.Error("expected Long description to mention 'Revoke a pending invitation'")
	}
}

// TestNewTeamRevokeInvitationCmd_Aliases verifies that aliases are set correctly.
func TestNewTeamRevokeInvitationCmd_Aliases(t *testing.T) {
	cmd := NewTeamRevokeInvitationCmd()

	expectedAliases := []string{"revoke-invite", "cancel-invitation", "cancel-invite"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for i, alias := range expectedAliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
			t.Errorf("expected alias %d to be %q, got %q", i, alias, cmd.Aliases[i])
		}
	}
}

// TestNewTeamRevokeInvitationCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewTeamRevokeInvitationCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewTeamRevokeInvitationCmd()

	// Verify examples are present
	examples := []string{
		"stackeye team revoke-invitation",
		"--id",
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

// TestNewTeamRevokeInvitationCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewTeamRevokeInvitationCmd_RunEIsSet(t *testing.T) {
	cmd := NewTeamRevokeInvitationCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewTeamRevokeInvitationCmd_HasFlags verifies flags are defined.
func TestNewTeamRevokeInvitationCmd_HasFlags(t *testing.T) {
	cmd := NewTeamRevokeInvitationCmd()

	// Check --id flag
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("expected --id flag to be defined")
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

// TestValidateRevokeInvitationFlags_ValidValues verifies valid flag values are accepted.
func TestValidateRevokeInvitationFlags_ValidValues(t *testing.T) {
	tests := []struct {
		name  string
		flags *teamRevokeInvitationFlags
	}{
		{
			name:  "valid id only",
			flags: &teamRevokeInvitationFlags{id: "abc123def456"},
		},
		{
			name:  "valid id with force",
			flags: &teamRevokeInvitationFlags{id: "abc123def456", force: true},
		},
		{
			name:  "uuid format id",
			flags: &teamRevokeInvitationFlags{id: "550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name:  "valid email only",
			flags: &teamRevokeInvitationFlags{email: "test@stackeye.io"},
		},
		{
			name:  "valid email with force",
			flags: &teamRevokeInvitationFlags{email: "test@stackeye.io", force: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRevokeInvitationFlags(tt.flags)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestValidateRevokeInvitationFlags_InvalidValues verifies invalid flag values are rejected.
func TestValidateRevokeInvitationFlags_InvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		flags       *teamRevokeInvitationFlags
		errContains string
	}{
		{
			name:        "neither id nor email",
			flags:       &teamRevokeInvitationFlags{},
			errContains: "either --id or --email is required",
		},
		{
			name:        "neither id nor email with force",
			flags:       &teamRevokeInvitationFlags{force: true},
			errContains: "either --id or --email is required",
		},
		{
			name:        "both id and email",
			flags:       &teamRevokeInvitationFlags{id: "abc123", email: "test@stackeye.io"},
			errContains: "cannot specify both --id and --email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRevokeInvitationFlags(tt.flags)
			if err == nil {
				t.Error("expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

// TestTeamRevokeInvitationTimeout verifies the timeout constant is reasonable.
func TestTeamRevokeInvitationTimeout(t *testing.T) {
	// Timeout should be at least 10 seconds for network operations
	if teamRevokeInvitationTimeout.Seconds() < 10 {
		t.Errorf("expected timeout to be at least 10 seconds, got %v", teamRevokeInvitationTimeout)
	}

	// Timeout should not exceed 60 seconds for a simple operation
	if teamRevokeInvitationTimeout.Seconds() > 60 {
		t.Errorf("expected timeout to be at most 60 seconds, got %v", teamRevokeInvitationTimeout)
	}
}
