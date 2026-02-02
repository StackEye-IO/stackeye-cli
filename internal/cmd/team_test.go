package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestNewTeamCmd(t *testing.T) {
	cmd := NewTeamCmd()

	if cmd.Use != "team" {
		t.Errorf("expected Use='team', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Manage team members" {
		t.Errorf("expected Short='Manage team members', got %q", cmd.Short)
	}
}

func TestNewTeamCmd_Long(t *testing.T) {
	cmd := NewTeamCmd()

	// Verify long description contains key information
	long := cmd.Long

	// Should explain what teams are
	if !strings.Contains(long, "members") {
		t.Error("expected Long description to mention members")
	}
	if !strings.Contains(long, "organization") {
		t.Error("expected Long description to mention organization")
	}
	if !strings.Contains(long, "role") {
		t.Error("expected Long description to mention role")
	}

	// Should describe available roles
	roles := []string{"owner", "admin", "member", "viewer"}
	for _, role := range roles {
		if !strings.Contains(long, role) {
			t.Errorf("expected Long description to mention %q role", role)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye team") {
		t.Error("expected Long description to contain example commands")
	}
}

func TestNewTeamCmd_Aliases(t *testing.T) {
	cmd := NewTeamCmd()

	expectedAliases := []string{"teams", "members"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewTeamCmd_Examples(t *testing.T) {
	cmd := NewTeamCmd()

	// Should have practical examples
	examples := []string{
		"team list",
		"team invite",
		"team update-role",
		"team remove",
		"team invitations",
		"team revoke-invitation",
	}

	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

func TestNewTeamCmd_Subcommands(t *testing.T) {
	cmd := NewTeamCmd()

	// Verify expected subcommands are registered
	expectedSubcommands := []string{"list", "invite", "update-role", "remove", "invitations", "revoke-invitation"}

	if len(cmd.Commands()) != len(expectedSubcommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubcommands), len(cmd.Commands()))
	}

	for _, name := range expectedSubcommands {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}
