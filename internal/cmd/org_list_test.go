// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewOrgListCmd verifies that the org list command is properly constructed.
func TestNewOrgListCmd(t *testing.T) {
	cmd := NewOrgListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "organization") {
		t.Error("expected Long description to mention 'organization'")
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

// TestNewOrgListCmd_HelpContainsRoles verifies that the help text explains user roles.
func TestNewOrgListCmd_HelpContainsRoles(t *testing.T) {
	cmd := NewOrgListCmd()

	// Verify all roles are documented
	roles := []string{"owner", "admin", "member", "viewer"}
	for _, role := range roles {
		if !strings.Contains(cmd.Long, role) {
			t.Errorf("expected Long description to document role %q", role)
		}
	}
}

// TestNewOrgListCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewOrgListCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewOrgListCmd()

	// Verify examples are present
	examples := []string{
		"stackeye org list",
		"-o json",
		"-o yaml",
	}
	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

// TestNewOrgListCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewOrgListCmd_RunEIsSet(t *testing.T) {
	cmd := NewOrgListCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}
