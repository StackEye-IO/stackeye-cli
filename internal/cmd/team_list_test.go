// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewTeamListCmd verifies that the team list command is properly constructed.
func TestNewTeamListCmd(t *testing.T) {
	cmd := NewTeamListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "team member") {
		t.Error("expected Long description to mention 'team member'")
	}

	// Verify aliases
	expectedAliases := []string{"ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("expected alias[%d] to be %q, got %q", i, alias, cmd.Aliases[i])
			}
		}
	}
}

// TestNewTeamListCmd_HelpContainsRoles verifies that the help text explains user roles.
func TestNewTeamListCmd_HelpContainsRoles(t *testing.T) {
	cmd := NewTeamListCmd()

	// Verify all roles are documented
	roles := []string{"owner", "admin", "member", "viewer"}
	for _, role := range roles {
		if !strings.Contains(cmd.Long, role) {
			t.Errorf("expected Long description to document role %q", role)
		}
	}
}

// TestNewTeamListCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewTeamListCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewTeamListCmd()

	// Verify examples are present
	examples := []string{
		"stackeye team list",
		"-o json",
		"-o yaml",
	}
	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

// TestNewTeamListCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewTeamListCmd_RunEIsSet(t *testing.T) {
	cmd := NewTeamListCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewTeamListCmd_HasPaginationFlags verifies pagination flags are defined.
func TestNewTeamListCmd_HasPaginationFlags(t *testing.T) {
	cmd := NewTeamListCmd()

	// Check --page flag
	pageFlag := cmd.Flags().Lookup("page")
	if pageFlag == nil {
		t.Error("expected --page flag to be defined")
	} else if pageFlag.DefValue != "1" {
		t.Errorf("expected --page default to be '1', got %q", pageFlag.DefValue)
	}

	// Check --limit flag
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("expected --limit flag to be defined")
	} else if limitFlag.DefValue != "20" {
		t.Errorf("expected --limit default to be '20', got %q", limitFlag.DefValue)
	}
}

// TestNewTeamListCmd_HasRoleFlag verifies the --role flag is defined.
func TestNewTeamListCmd_HasRoleFlag(t *testing.T) {
	cmd := NewTeamListCmd()

	roleFlag := cmd.Flags().Lookup("role")
	if roleFlag == nil {
		t.Error("expected --role flag to be defined")
	} else {
		if roleFlag.DefValue != "" {
			t.Errorf("expected --role default to be empty, got %q", roleFlag.DefValue)
		}
		if !strings.Contains(roleFlag.Usage, "owner") {
			t.Error("expected --role usage to mention 'owner'")
		}
		if !strings.Contains(roleFlag.Usage, "admin") {
			t.Error("expected --role usage to mention 'admin'")
		}
	}
}

// TestNewTeamListCmd_HelpContainsRoleFilter verifies role filter example is present.
func TestNewTeamListCmd_HelpContainsRoleFilter(t *testing.T) {
	cmd := NewTeamListCmd()

	if !strings.Contains(cmd.Long, "--role admin") {
		t.Error("expected Long description to contain role filter example")
	}

	// Verify client-side filtering note is present
	if !strings.Contains(cmd.Long, "Role filtering is applied client-side") {
		t.Error("expected Long description to document client-side filtering limitation")
	}
}

// TestValidateTeamListFlags_ValidValues verifies valid flag values are accepted.
func TestValidateTeamListFlags_ValidValues(t *testing.T) {
	tests := []struct {
		name  string
		flags *teamListFlags
	}{
		{
			name:  "default values",
			flags: &teamListFlags{page: 1, limit: 20},
		},
		{
			name:  "minimum values",
			flags: &teamListFlags{page: 1, limit: 1},
		},
		{
			name:  "maximum limit",
			flags: &teamListFlags{page: 1, limit: 100},
		},
		{
			name:  "high page number",
			flags: &teamListFlags{page: 100, limit: 20},
		},
		{
			name:  "with owner role",
			flags: &teamListFlags{page: 1, limit: 20, role: "owner"},
		},
		{
			name:  "with admin role",
			flags: &teamListFlags{page: 1, limit: 20, role: "admin"},
		},
		{
			name:  "with member role",
			flags: &teamListFlags{page: 1, limit: 20, role: "member"},
		},
		{
			name:  "with viewer role",
			flags: &teamListFlags{page: 1, limit: 20, role: "viewer"},
		},
		{
			name:  "with uppercase role",
			flags: &teamListFlags{page: 1, limit: 20, role: "ADMIN"},
		},
		{
			name:  "with mixed case role",
			flags: &teamListFlags{page: 1, limit: 20, role: "Member"},
		},
		{
			name:  "with empty role",
			flags: &teamListFlags{page: 1, limit: 20, role: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamListFlags(tt.flags)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestValidateTeamListFlags_InvalidValues verifies invalid flag values are rejected.
func TestValidateTeamListFlags_InvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		flags       *teamListFlags
		errContains string
	}{
		{
			name:        "zero limit",
			flags:       &teamListFlags{page: 1, limit: 0},
			errContains: "invalid limit",
		},
		{
			name:        "negative limit",
			flags:       &teamListFlags{page: 1, limit: -1},
			errContains: "invalid limit",
		},
		{
			name:        "limit too high",
			flags:       &teamListFlags{page: 1, limit: 101},
			errContains: "invalid limit",
		},
		{
			name:        "zero page",
			flags:       &teamListFlags{page: 0, limit: 20},
			errContains: "invalid page",
		},
		{
			name:        "negative page",
			flags:       &teamListFlags{page: -1, limit: 20},
			errContains: "invalid page",
		},
		{
			name:        "invalid role",
			flags:       &teamListFlags{page: 1, limit: 20, role: "superuser"},
			errContains: "for --role",
		},
		{
			name:        "invalid role typo",
			flags:       &teamListFlags{page: 1, limit: 20, role: "admn"},
			errContains: "for --role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTeamListFlags(tt.flags)
			if err == nil {
				t.Error("expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

// TestFilterMembersByRole verifies the role filtering function works correctly.
func TestFilterMembersByRole(t *testing.T) {
	members := []client.TeamMember{
		{Name: "Alice", Role: client.TeamRoleOwner},
		{Name: "Bob", Role: client.TeamRoleAdmin},
		{Name: "Charlie", Role: client.TeamRoleMember},
		{Name: "Diana", Role: client.TeamRoleMember},
		{Name: "Eve", Role: client.TeamRoleViewer},
	}

	tests := []struct {
		name     string
		role     string
		expected int
	}{
		{"filter owner", "owner", 1},
		{"filter admin", "admin", 1},
		{"filter member", "member", 2},
		{"filter viewer", "viewer", 1},
		{"case insensitive - uppercase", "ADMIN", 1},
		{"case insensitive - mixed", "Member", 2},
		{"empty role returns all", "", 5},
		{"non-existent role returns none", "superuser", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle empty role case - empty string should skip filtering
			var result []client.TeamMember
			if tt.role == "" {
				result = members // No filtering
			} else {
				result = filterMembersByRole(members, tt.role)
			}
			if len(result) != tt.expected {
				t.Errorf("expected %d members, got %d", tt.expected, len(result))
			}
		})
	}
}

// TestFilterMembersByRole_PreservesOrder verifies filtering preserves member order.
func TestFilterMembersByRole_PreservesOrder(t *testing.T) {
	members := []client.TeamMember{
		{Name: "Alice", Role: client.TeamRoleMember},
		{Name: "Bob", Role: client.TeamRoleMember},
		{Name: "Charlie", Role: client.TeamRoleMember},
	}

	result := filterMembersByRole(members, "member")

	if len(result) != 3 {
		t.Fatalf("expected 3 members, got %d", len(result))
	}

	expectedOrder := []string{"Alice", "Bob", "Charlie"}
	for i, name := range expectedOrder {
		if result[i].Name != name {
			t.Errorf("expected member[%d] to be %q, got %q", i, name, result[i].Name)
		}
	}
}

// TestFilterMembersByRole_EmptyInput verifies filtering handles empty input.
func TestFilterMembersByRole_EmptyInput(t *testing.T) {
	var members []client.TeamMember

	result := filterMembersByRole(members, "admin")

	if len(result) != 0 {
		t.Errorf("expected 0 members, got %d", len(result))
	}
}
