// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
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
