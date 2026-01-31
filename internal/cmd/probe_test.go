package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestNewProbeCmd(t *testing.T) {
	cmd := NewProbeCmd()

	if cmd.Use != "probe" {
		t.Errorf("expected Use='probe', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Manage monitoring probes" {
		t.Errorf("expected Short='Manage monitoring probes', got %q", cmd.Short)
	}
}

func TestNewProbeCmd_Long(t *testing.T) {
	cmd := NewProbeCmd()

	// Verify long description contains key information
	long := cmd.Long

	// Should explain what probes are
	if !strings.Contains(long, "monitoring") {
		t.Error("expected Long description to mention monitoring")
	}
	if !strings.Contains(long, "endpoint") {
		t.Error("expected Long description to mention endpoint")
	}

	// Should list available subcommands
	subcommands := []string{"list", "get", "create", "update", "delete", "pause", "resume", "test", "history", "stats"}
	for _, sub := range subcommands {
		if !strings.Contains(long, sub) {
			t.Errorf("expected Long description to mention %q subcommand", sub)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye probe") {
		t.Error("expected Long description to contain example commands")
	}
}

func TestNewProbeCmd_Aliases(t *testing.T) {
	cmd := NewProbeCmd()

	expectedAliases := []string{"probes", "monitor", "monitors"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewProbeCmd_Examples(t *testing.T) {
	cmd := NewProbeCmd()

	// Should have practical examples
	examples := []string{
		"probe list",
		"probe get",
		"probe create",
		"probe pause",
		"probe resume",
		"probe test",
		"probe stats",
	}

	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}

func TestNewProbeCmd_Subcommands(t *testing.T) {
	cmd := NewProbeCmd()

	// Verify expected subcommands are registered
	expectedSubcommands := []string{"list", "get", "create", "wizard", "update", "delete", "pause", "resume", "test", "history", "stats", "link-channel", "deps"}

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
