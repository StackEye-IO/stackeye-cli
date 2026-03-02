// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
)

// TestNewPrivateRegionRevokeCmd verifies the revoke command is correctly configured.
func TestNewPrivateRegionRevokeCmd(t *testing.T) {
	cmd := NewPrivateRegionRevokeCmd()

	if cmd.Use != "revoke" {
		t.Errorf("expected Use='revoke', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewPrivateRegionRevokeCmd_FlagsExist verifies --region-id and --key-id are declared.
func TestNewPrivateRegionRevokeCmd_FlagsExist(t *testing.T) {
	cmd := NewPrivateRegionRevokeCmd()

	if cmd.Flags().Lookup("region-id") == nil {
		t.Error("expected flag --region-id to be defined")
	}
	if cmd.Flags().Lookup("key-id") == nil {
		t.Error("expected flag --key-id to be defined")
	}
}

// TestNewPrivateRegionRevokeCmd_FlagShorthands verifies shorthands -r and -k.
func TestNewPrivateRegionRevokeCmd_FlagShorthands(t *testing.T) {
	cmd := NewPrivateRegionRevokeCmd()

	regionFlag := cmd.Flags().Lookup("region-id")
	if regionFlag == nil {
		t.Fatal("--region-id flag not found")
	}
	if regionFlag.Shorthand != "r" {
		t.Errorf("expected --region-id shorthand 'r', got %q", regionFlag.Shorthand)
	}

	keyFlag := cmd.Flags().Lookup("key-id")
	if keyFlag == nil {
		t.Fatal("--key-id flag not found")
	}
	if keyFlag.Shorthand != "k" {
		t.Errorf("expected --key-id shorthand 'k', got %q", keyFlag.Shorthand)
	}
}

// TestNewPrivateRegionRevokeCmd_Long verifies example commands in Long.
func TestNewPrivateRegionRevokeCmd_Long(t *testing.T) {
	cmd := NewPrivateRegionRevokeCmd()

	if !strings.Contains(cmd.Long, "stackeye private-region revoke") {
		t.Error("expected Long description to contain example commands")
	}
	if !strings.Contains(cmd.Long, "--dry-run") {
		t.Error("expected Long description to mention --dry-run")
	}
}
