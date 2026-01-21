// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewOrgGetCmd verifies that the org get command is properly constructed.
func TestNewOrgGetCmd(t *testing.T) {
	cmd := NewOrgGetCmd()

	if cmd.Use != "get [id|slug]" {
		t.Errorf("expected Use to be 'get [id|slug]', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "organization") {
		t.Error("expected Long description to mention 'organization'")
	}

	// Verify aliases
	expectedAliases := []string{"show", "info"}
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

// TestNewOrgGetCmd_HelpContainsRoles verifies that the help text explains user roles.
func TestNewOrgGetCmd_HelpContainsRoles(t *testing.T) {
	cmd := NewOrgGetCmd()

	// Verify all roles are documented
	roles := []string{"owner", "admin", "member", "viewer"}
	for _, role := range roles {
		if !strings.Contains(cmd.Long, role) {
			t.Errorf("expected Long description to document role %q", role)
		}
	}
}

// TestNewOrgGetCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewOrgGetCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewOrgGetCmd()

	// Verify examples are present
	examples := []string{
		"stackeye org get",
		"-o json",
		"-o yaml",
		"acme-corp",
	}
	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

// TestNewOrgGetCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewOrgGetCmd_RunEIsSet(t *testing.T) {
	cmd := NewOrgGetCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewOrgGetCmd_AcceptsOptionalArg verifies that the command accepts 0 or 1 arguments.
func TestNewOrgGetCmd_AcceptsOptionalArg(t *testing.T) {
	cmd := NewOrgGetCmd()

	// Should accept 0 args
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("expected command to accept 0 args, got error: %v", err)
	}

	// Should accept 1 arg
	err = cmd.Args(cmd, []string{"org-id"})
	if err != nil {
		t.Errorf("expected command to accept 1 arg, got error: %v", err)
	}

	// Should reject 2+ args
	err = cmd.Args(cmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("expected command to reject 2 args, but it accepted them")
	}
}

// TestNewOrgGetCmd_HelpContainsPlanLimits verifies that plan limits are documented.
func TestNewOrgGetCmd_HelpContainsPlanLimits(t *testing.T) {
	cmd := NewOrgGetCmd()

	// Verify plan limit concepts are documented
	concepts := []string{"Monitor", "Team member"}
	for _, concept := range concepts {
		if !strings.Contains(cmd.Long, concept) {
			t.Errorf("expected Long description to document %q", concept)
		}
	}
}
