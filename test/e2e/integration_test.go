//go:build integration
// +build integration

// Package e2e provides integration tests that run against the live dev API.
// These tests require the user to be authenticated (run `stackeye login` first).
//
// Run with: STACKEYE_E2E_LIVE=true go test ./test/e2e/... -tags=integration -v
// Or:       make test-integration
package e2e

import (
	"testing"

	"github.com/StackEye-IO/stackeye-cli/test/e2e/testutil"
)

// TestLive_ProbeList tests the probe list command against the live API.
func TestLive_ProbeList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default table output", func(t *testing.T) {
		result := env.RunSuccess("probe", "list")

		// Verify output format, not specific fixture data
		// Real API may return 0+ probes, but header should be present
		env.AssertContains(result, "NAME")
	})

	t.Run("json output format", func(t *testing.T) {
		result := env.RunSuccess("probe", "list", "--output", "json")

		// Should be valid JSON (even if empty array)
		var data any
		env.AssertJSONOutput(result, &data)
	})

	t.Run("yaml output format", func(t *testing.T) {
		result := env.RunSuccess("probe", "list", "--output", "yaml")

		// YAML output - if probes exist, should contain name field
		// If no probes, output may be minimal but should not error
		_ = result // Just verify no error
	})
}

// TestLive_AlertList tests the alert list command against the live API.
func TestLive_AlertList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default table output", func(t *testing.T) {
		result := env.RunSuccess("alert", "list")

		// Live API may return 0 alerts - verify either header or empty message
		hasHeader := result.Contains("SEVERITY")
		hasEmptyMessage := result.Contains("No alerts found")
		if !hasHeader && !hasEmptyMessage {
			t.Errorf("expected either table header or empty message, got: %s", result.Stdout)
		}
	})

	t.Run("json output format", func(t *testing.T) {
		result := env.RunSuccess("alert", "list", "--output", "json")

		// Should be valid JSON (even if empty array)
		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestLive_ChannelList tests the channel list command against the live API.
func TestLive_ChannelList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default table output", func(t *testing.T) {
		result := env.RunSuccess("channel", "list")

		// Live API may return 0 channels - verify either header or empty message
		hasHeader := result.Contains("NAME")
		hasEmptyMessage := result.Contains("No channels found")
		if !hasHeader && !hasEmptyMessage {
			t.Errorf("expected either table header or empty message, got: %s", result.Stdout)
		}
	})

	t.Run("json output format", func(t *testing.T) {
		result := env.RunSuccess("channel", "list", "--output", "json")

		// Should be valid JSON (even if empty array)
		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestLive_ContextList tests the context list command.
// Note: This doesn't make API calls - it reads the local config.
func TestLive_ContextList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	result := env.RunSuccess("context", "list")

	// Should show at least one context (the current one)
	env.AssertContains(result, "NAME")
	// Should have a current context marker
	env.AssertContains(result, "*")
}

// TestLive_ContextCurrent tests the context current command.
func TestLive_ContextCurrent(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	result := env.RunSuccess("context", "current")

	// Should output the current context name
	// The output is non-empty since we validated a context exists
	lines := result.Lines()
	if len(lines) == 0 {
		t.Error("expected context current to output current context name")
	}
}

// TestLive_Version tests the version command.
// This doesn't require API access, but verifies the CLI is built correctly.
func TestLive_Version(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	result := env.RunSuccess("version")

	// Should contain version info
	env.AssertContains(result, "stackeye")
}

// TestLive_Help tests the help command.
func TestLive_Help(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	result := env.RunSuccess("help")

	// Should list available commands
	env.AssertContains(result, "probe")
	env.AssertContains(result, "alert")
	env.AssertContains(result, "channel")
	env.AssertContains(result, "context")
}
