package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestNewProbeDepsCmd(t *testing.T) {
	cmd := NewProbeDepsCmd()

	if cmd.Use != "deps" {
		t.Errorf("expected Use='deps', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Manage probe dependencies" {
		t.Errorf("expected Short='Manage probe dependencies', got %q", cmd.Short)
	}
}

func TestNewProbeDepsCmd_Long(t *testing.T) {
	cmd := NewProbeDepsCmd()

	long := cmd.Long

	// Should explain hierarchical alerting concept
	if !strings.Contains(long, "hierarchical alerting") {
		t.Error("expected Long description to mention hierarchical alerting")
	}

	// Should explain parent/child relationship
	if !strings.Contains(long, "parent") {
		t.Error("expected Long description to mention parent")
	}

	// Should explain the UNREACHABLE status
	if !strings.Contains(long, "UNREACHABLE") {
		t.Error("expected Long description to mention UNREACHABLE status")
	}

	// Should explain alert suppression
	if !strings.Contains(long, "suppress") {
		t.Error("expected Long description to mention alert suppression")
	}

	// Should list available subcommands
	subcommands := []string{"list", "add", "remove", "clear", "tree", "wizard"}
	for _, sub := range subcommands {
		if !strings.Contains(long, sub) {
			t.Errorf("expected Long description to mention %q subcommand", sub)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye probe deps") {
		t.Error("expected Long description to contain example commands")
	}
}

func TestNewProbeDepsCmd_Aliases(t *testing.T) {
	cmd := NewProbeDepsCmd()

	expectedAliases := []string{"dependencies", "dependency", "dep"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewProbeDepsCmd_Examples(t *testing.T) {
	cmd := NewProbeDepsCmd()

	// Should have practical examples for each operation
	examples := []string{
		"deps list",
		"deps add",
		"deps remove",
		"deps clear",
		"deps tree",
		"deps wizard",
	}

	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

func TestNewProbeDepsCmd_CommonPatterns(t *testing.T) {
	cmd := NewProbeDepsCmd()

	// Should explain common dependency patterns
	patterns := []string{
		"Database",
		"Load Balancer",
		"Router",
	}

	for _, pattern := range patterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long description to explain %q dependency pattern", pattern)
		}
	}
}
