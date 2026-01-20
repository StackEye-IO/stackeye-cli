package cmd

import (
	"slices"
	"strings"
	"testing"
)

func TestNewAlertCmd(t *testing.T) {
	cmd := NewAlertCmd()

	if cmd.Use != "alert" {
		t.Errorf("expected Use='alert', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Manage monitoring alerts" {
		t.Errorf("expected Short='Manage monitoring alerts', got %q", cmd.Short)
	}
}

func TestNewAlertCmd_Long(t *testing.T) {
	cmd := NewAlertCmd()

	long := cmd.Long

	// Should explain what alerts are
	if !strings.Contains(long, "triggered") {
		t.Error("expected Long description to explain when alerts are triggered")
	}
	if !strings.Contains(long, "probe") {
		t.Error("expected Long description to mention probes")
	}

	// Should document alert states
	states := []string{"active", "acknowledged", "resolved"}
	for _, state := range states {
		if !strings.Contains(long, state) {
			t.Errorf("expected Long description to document %q state", state)
		}
	}

	// Should document severity levels
	severities := []string{"critical", "warning", "info"}
	for _, sev := range severities {
		if !strings.Contains(long, sev) {
			t.Errorf("expected Long description to document %q severity", sev)
		}
	}

	// Should list available subcommands
	subcommands := []string{"list", "get", "ack", "resolve", "history"}
	for _, sub := range subcommands {
		if !strings.Contains(long, sub) {
			t.Errorf("expected Long description to mention %q subcommand", sub)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye alert") {
		t.Error("expected Long description to contain example commands")
	}
}

func TestNewAlertCmd_Aliases(t *testing.T) {
	cmd := NewAlertCmd()

	expectedAliases := []string{"alerts", "alerting"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	for _, expected := range expectedAliases {
		if !slices.Contains(cmd.Aliases, expected) {
			t.Errorf("expected alias %q not found", expected)
		}
	}
}

func TestNewAlertCmd_Examples(t *testing.T) {
	cmd := NewAlertCmd()

	// Should have practical examples
	examples := []string{
		"alert list",
		"alert get",
		"alert ack",
		"alert resolve",
		"alert history",
	}

	for _, example := range examples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("expected Long description to contain example %q", example)
		}
	}
}
