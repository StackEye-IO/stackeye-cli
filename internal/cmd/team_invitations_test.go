// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewTeamInvitationsCmd verifies that the team invitations command is properly constructed.
func TestNewTeamInvitationsCmd(t *testing.T) {
	cmd := NewTeamInvitationsCmd()

	if cmd.Use != "invitations" {
		t.Errorf("expected Use to be 'invitations', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "pending invitation") {
		t.Error("expected Long description to mention 'pending invitation'")
	}
}

// TestNewTeamInvitationsCmd_HasAliases verifies that command aliases are defined.
func TestNewTeamInvitationsCmd_HasAliases(t *testing.T) {
	cmd := NewTeamInvitationsCmd()

	expectedAliases := []string{"invites", "inv"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
		return
	}

	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("expected alias[%d] to be %q, got %q", i, alias, cmd.Aliases[i])
		}
	}
}

// TestNewTeamInvitationsCmd_HelpContainsRoles verifies that the help text explains assignable roles.
func TestNewTeamInvitationsCmd_HelpContainsRoles(t *testing.T) {
	cmd := NewTeamInvitationsCmd()

	// Verify roles that can be assigned via invitation are documented
	roles := []string{"admin", "member", "viewer"}
	for _, role := range roles {
		if !strings.Contains(cmd.Long, role) {
			t.Errorf("expected Long description to document role %q", role)
		}
	}
}

// TestNewTeamInvitationsCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewTeamInvitationsCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewTeamInvitationsCmd()

	// Verify examples are present
	examples := []string{
		"stackeye team invitations",
		"-o json",
		"-o yaml",
		"-o wide",
	}
	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

// TestNewTeamInvitationsCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewTeamInvitationsCmd_RunEIsSet(t *testing.T) {
	cmd := NewTeamInvitationsCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewTeamInvitationsCmd_NoPaginationFlags verifies pagination flags are NOT defined.
// The backend returns all invitations without pagination.
func TestNewTeamInvitationsCmd_NoPaginationFlags(t *testing.T) {
	cmd := NewTeamInvitationsCmd()

	// Verify --page flag is NOT defined (unlike team list)
	pageFlag := cmd.Flags().Lookup("page")
	if pageFlag != nil {
		t.Error("expected --page flag to NOT be defined for invitations")
	}

	// Verify --limit flag is NOT defined
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag != nil {
		t.Error("expected --limit flag to NOT be defined for invitations")
	}
}

// TestNewTeamInvitationsCmd_MentionsExpiry verifies that the help mentions invitation expiry.
func TestNewTeamInvitationsCmd_MentionsExpiry(t *testing.T) {
	cmd := NewTeamInvitationsCmd()

	if !strings.Contains(cmd.Long, "expire") {
		t.Error("expected Long description to mention invitation expiry")
	}
}
