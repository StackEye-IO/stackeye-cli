// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"strings"
	"testing"
)

func TestNewEnrollmentKeyCreateCmd(t *testing.T) {
	cmd := NewEnrollmentKeyCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create")
	}

	for _, name := range []string{"name", "capability", "mode", "max-uses", "ttl-seconds", "unbounded"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected --%s flag to be defined", name)
		}
	}
}

func TestEnrollmentKeyCreateCmd_Parent(t *testing.T) {
	parentCmd := NewEnrollmentKeyCmd()

	var foundCreate bool
	for _, sub := range parentCmd.Commands() {
		if sub.Use == "create" {
			foundCreate = true
		}
	}
	if !foundCreate {
		t.Error("expected 'create' subcommand to be registered under enrollment-key command")
	}
}

func TestRunEnrollmentKeyCreate_RequiresClient(t *testing.T) {
	flags := &enrollmentKeyCreateFlags{name: "Test Key"}

	// No real backend/config exists in the test environment: this fails either
	// at client-init ("no client configured") or, if a sibling test in this
	// package left a fake authenticated config in place, at the HTTP call
	// itself — either way, no error is unexpected here.
	err := runEnrollmentKeyCreate(context.Background(), flags)
	if err == nil {
		t.Fatal("expected an error with no real backend configured, got nil")
	}
}

func TestRunEnrollmentKeyCreate_InvalidCapability(t *testing.T) {
	flags := &enrollmentKeyCreateFlags{capability: []string{"bogus_capability"}}

	err := runEnrollmentKeyCreate(context.Background(), flags)
	if err == nil {
		t.Fatal("expected error for invalid capability, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --capability") {
		t.Errorf("error should mention invalid --capability, got: %v", err)
	}
}

func TestRunEnrollmentKeyCreate_InvalidMode(t *testing.T) {
	flags := &enrollmentKeyCreateFlags{mode: "turbo"}

	err := runEnrollmentKeyCreate(context.Background(), flags)
	if err == nil {
		t.Fatal("expected error for invalid mode, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --mode") {
		t.Errorf("error should mention invalid --mode, got: %v", err)
	}
}

func TestRunEnrollmentKeyCreate_UnboundedConflictsWithMaxUses(t *testing.T) {
	flags := &enrollmentKeyCreateFlags{unbounded: true, maxUses: 10}

	err := runEnrollmentKeyCreate(context.Background(), flags)
	if err == nil {
		t.Fatal("expected error for --unbounded + --max-uses conflict, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Errorf("error should mention the conflict, got: %v", err)
	}
}

func TestRunEnrollmentKeyCreate_UnboundedConflictsWithTTL(t *testing.T) {
	flags := &enrollmentKeyCreateFlags{unbounded: true, ttlSeconds: 3600}

	err := runEnrollmentKeyCreate(context.Background(), flags)
	if err == nil {
		t.Fatal("expected error for --unbounded + --ttl-seconds conflict, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Errorf("error should mention the conflict, got: %v", err)
	}
}

func TestEnrollmentKeyCreateCmd_ValidModes(t *testing.T) {
	for _, mode := range []string{"", "standard", "fleet"} {
		flags := &enrollmentKeyCreateFlags{mode: mode}
		err := runEnrollmentKeyCreate(context.Background(), flags)
		// Fails at the API-client stage (none configured in tests), never at mode validation.
		if err != nil && strings.Contains(err.Error(), "invalid --mode") {
			t.Errorf("mode %q should be valid, got: %v", mode, err)
		}
	}
}
