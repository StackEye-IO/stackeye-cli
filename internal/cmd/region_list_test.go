// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewRegionListCmd verifies that the region list command is properly constructed.
func TestNewRegionListCmd(t *testing.T) {
	cmd := NewRegionListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify Long description contains key information
	if !strings.Contains(cmd.Long, "region") {
		t.Error("expected Long description to mention 'region'")
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

// TestNewRegionListCmd_HelpContainsColumns verifies that the help text explains output columns.
func TestNewRegionListCmd_HelpContainsColumns(t *testing.T) {
	cmd := NewRegionListCmd()

	// Verify all output columns are documented
	columns := []string{"CODE", "NAME", "DISPLAY", "COUNTRY", "CONTINENT"}
	for _, column := range columns {
		if !strings.Contains(cmd.Long, column) {
			t.Errorf("expected Long description to document column %q", column)
		}
	}
}

// TestNewRegionListCmd_HelpContainsExamples verifies that usage examples are provided.
func TestNewRegionListCmd_HelpContainsExamples(t *testing.T) {
	cmd := NewRegionListCmd()

	// Verify examples are present
	examples := []string{
		"stackeye region list",
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

// TestNewRegionListCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewRegionListCmd_RunEIsSet(t *testing.T) {
	cmd := NewRegionListCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewRegionListCmd_AuthRequirementNote verifies that the help text mentions authentication is required.
func TestNewRegionListCmd_AuthRequirementNote(t *testing.T) {
	cmd := NewRegionListCmd()

	// Verify authentication requirement note is present
	if !strings.Contains(cmd.Long, "Requires authentication") {
		t.Error("expected Long description to note that authentication is required")
	}
}
