// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewAgentCmd verifies the parent command is correctly configured.
func TestNewAgentCmd(t *testing.T) {
	cmd := NewAgentCmd()

	if cmd.Use != "agent" {
		t.Errorf("expected Use='agent', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE != nil {
		t.Error("parent command should not have RunE (it has subcommands)")
	}
}

// TestNewAgentCmd_SubcommandsRegistered verifies all subcommands are present.
func TestNewAgentCmd_SubcommandsRegistered(t *testing.T) {
	cmd := NewAgentCmd()

	want := []string{"list", "get", "register"}
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

// TestNewAgentCmd_Long verifies the Long description contains key concepts.
func TestNewAgentCmd_Long(t *testing.T) {
	cmd := NewAgentCmd()
	long := cmd.Long

	keywords := []string{"agent", "metrics", "api key"}
	for _, kw := range keywords {
		if !strings.Contains(strings.ToLower(long), kw) {
			t.Errorf("expected Long description to mention %q", kw)
		}
	}
}

// TestNewAgentCmd_Aliases verifies at least one alias is set.
func TestNewAgentCmd_Aliases(t *testing.T) {
	cmd := NewAgentCmd()
	if len(cmd.Aliases) == 0 {
		t.Error("expected at least one alias")
	}
}
