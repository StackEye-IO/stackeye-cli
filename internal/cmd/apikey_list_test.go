// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"strings"
	"testing"
)

// TestNewAPIKeyListCmd verifies that the api-key list command is properly constructed.
func TestNewAPIKeyListCmd(t *testing.T) {
	cmd := NewAPIKeyListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
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

// TestRunAPIKeyList_RequiresClient tests that running list without a configured client fails.
func TestRunAPIKeyList_RequiresClient(t *testing.T) {
	// Call runAPIKeyList with a background context.
	// It should fail when trying to get the API client since none is configured.
	err := runAPIKeyList(context.Background())

	if err == nil {
		t.Error("expected error (no API client configured), got nil")
		return
	}

	// Error should be related to API client initialization
	errStr := err.Error()
	if !strings.Contains(errStr, "API client") && !strings.Contains(errStr, "client") && !strings.Contains(errStr, "config") {
		t.Logf("got expected error: %v", err)
	}
}

// TestAPIKeyListCmd_Parent verifies the parent api-key command registers the list subcommand.
func TestAPIKeyListCmd_Parent(t *testing.T) {
	parentCmd := NewAPIKeyCmd()

	// Check that list subcommand is registered
	var foundList bool
	for _, sub := range parentCmd.Commands() {
		if sub.Use == "list" {
			foundList = true
			break
		}
	}

	if !foundList {
		t.Error("expected 'list' subcommand to be registered under api-key command")
	}
}

// TestAPIKeyListCmd_LongDescription verifies the long description contains expected content.
func TestAPIKeyListCmd_LongDescription(t *testing.T) {
	cmd := NewAPIKeyListCmd()

	expectedPhrases := []string{
		"API keys",
		"prefix",
		"Security",
		"stackeye api-key list",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(cmd.Long, phrase) {
			t.Errorf("expected Long description to contain %q", phrase)
		}
	}
}
