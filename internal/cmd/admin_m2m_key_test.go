// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"slices"
	"strings"
	"testing"
)

// TestNewAdminM2MKeyCmd verifies that the m2m-key parent command is properly constructed.
func TestNewAdminM2MKeyCmd(t *testing.T) {
	cmd := NewAdminM2MKeyCmd()

	if cmd.Use != "m2m-key" {
		t.Errorf("expected Use='m2m-key', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Manage machine-to-machine authentication keys" {
		t.Errorf("expected Short='Manage machine-to-machine authentication keys', got %q", cmd.Short)
	}
}

// TestNewAdminM2MKeyCmd_Long verifies the Long description contains key information.
func TestNewAdminM2MKeyCmd_Long(t *testing.T) {
	cmd := NewAdminM2MKeyCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"machine-to-machine",
		"M2M",
		"authentication",
		"service-to-service",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin m2m-key list") {
		t.Error("expected Long description to contain example commands")
	}
}

// TestNewAdminM2MKeyCmd_Aliases verifies that aliases are set correctly.
func TestNewAdminM2MKeyCmd_Aliases(t *testing.T) {
	cmd := NewAdminM2MKeyCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected aliases to be set")
	}

	expectedAliases := []string{"m2m", "m2mkey"}
	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected '%s' alias to be set", expected)
		}
	}
}

// TestNewAdminM2MKeyCmd_HasListSubcommand verifies that list subcommand is registered.
func TestNewAdminM2MKeyCmd_HasListSubcommand(t *testing.T) {
	cmd := NewAdminM2MKeyCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected m2m-key command to have at least 1 subcommand (list)")
	}

	// Verify list subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'list' subcommand to be registered")
	}
}

// TestNewAdminCmd_HasM2MKeySubcommand verifies that m2m-key subcommand is registered.
func TestNewAdminCmd_HasM2MKeySubcommand(t *testing.T) {
	cmd := NewAdminCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected admin command to have subcommands")
	}

	// Verify m2m-key subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "m2m-key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'm2m-key' subcommand to be registered under admin")
	}
}
