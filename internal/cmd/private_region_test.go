// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewPrivateRegionCmd verifies the parent command is correctly configured.
func TestNewPrivateRegionCmd(t *testing.T) {
	cmd := NewPrivateRegionCmd()

	if cmd.Use != "private-region" {
		t.Errorf("expected Use='private-region', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE != nil {
		t.Error("parent command should not have RunE (it has subcommands)")
	}
}

// TestNewPrivateRegionCmd_SubcommandsRegistered verifies all subcommands are present.
func TestNewPrivateRegionCmd_SubcommandsRegistered(t *testing.T) {
	cmd := NewPrivateRegionCmd()

	want := []string{"list", "get", "create", "rotate-key", "revoke"}
	for _, name := range want {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Use == name || strings.HasPrefix(sub.Use, name+" ") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

// TestNewPrivateRegionCmd_Long verifies the Long description contains key concepts.
func TestNewPrivateRegionCmd_Long(t *testing.T) {
	cmd := NewPrivateRegionCmd()
	long := cmd.Long

	keywords := []string{"private", "region", "bootstrap", "appliance"}
	for _, kw := range keywords {
		if !strings.Contains(strings.ToLower(long), kw) {
			t.Errorf("expected Long description to mention %q", kw)
		}
	}
}

// TestNewPrivateRegionCmd_Aliases verifies at least one alias is set.
func TestNewPrivateRegionCmd_Aliases(t *testing.T) {
	cmd := NewPrivateRegionCmd()
	if len(cmd.Aliases) == 0 {
		t.Error("expected at least one alias")
	}
}
