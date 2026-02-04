//go:build integration
// +build integration

// Package e2e provides integration tests that run against the live dev API.
// These tests require the user to be authenticated (run `stackeye login` first).
//
// Run with: STACKEYE_E2E_LIVE=true go test ./test/e2e/... -tags=integration -v
// Or:       make test-integration
package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/test/e2e/testutil"
	"github.com/stretchr/testify/require"
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

// TestLive_ProbeLifecycle tests creating, getting, and deleting a probe against the live API.
func TestLive_ProbeLifecycle(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	// Create a probe
	var probeID string
	t.Run("create probe", func(t *testing.T) {
		result := env.RunSuccess(
			"probe", "create",
			"--name", "Integration Test Probe",
			"--check-type", "http",
			"--url", "https://httpbin.org/status/200",
			"--interval", "300",
			"--output", "json",
		)

		var data map[string]interface{}
		env.AssertJSONOutput(result, &data)

		// Extract probe ID from JSON output
		id, ok := data["id"].(string)
		require.True(t, ok, "expected probe ID in response, got: %v", data)
		require.NotEmpty(t, id, "probe ID should not be empty")
		probeID = id
	})

	// Skip remaining subtests if creation failed
	if probeID == "" {
		t.Fatal("probe creation failed, skipping lifecycle tests")
	}

	// Cleanup: delete probe when test completes
	defer func() {
		_ = env.Run("probe", "delete", probeID, "--yes")
	}()

	t.Run("get created probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "get", probeID, "--output", "json")

		var data map[string]interface{}
		env.AssertJSONOutput(result, &data)

		// Verify the probe name matches
		name, _ := data["name"].(string)
		require.Equal(t, "Integration Test Probe", name)
	})

	t.Run("get created probe table output", func(t *testing.T) {
		result := env.RunSuccess("probe", "get", probeID)

		env.AssertContains(result, "Integration Test Probe")
	})

	t.Run("pause probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "pause", probeID, "--yes")

		env.AssertContains(result, "Paused")
	})

	t.Run("resume probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "resume", probeID, "--yes")

		env.AssertContains(result, "Resumed")
	})

	t.Run("delete probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "delete", probeID, "--yes")

		// Verify deletion succeeded
		_ = result

		// Verify probe is gone
		getResult := env.Run("probe", "get", probeID)
		require.True(t, getResult.Failed(), "expected get to fail after deletion")
	})
}

// TestLive_ChannelLifecycle tests creating, getting, and deleting a channel against the live API.
func TestLive_ChannelLifecycle(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	var channelID string
	t.Run("create email channel", func(t *testing.T) {
		result := env.RunSuccess(
			"channel", "create",
			"--name", "Integration Test Channel",
			"--type", "email",
			"--email", "integration-test@stackeye.io",
			"--output", "json",
		)

		var data map[string]interface{}
		env.AssertJSONOutput(result, &data)

		// Extract channel ID - may be nested under "channel" key
		channelData := data
		if nested, ok := data["channel"].(map[string]interface{}); ok {
			channelData = nested
		}
		id, ok := channelData["id"].(string)
		require.True(t, ok, "expected channel ID in response, got: %v", data)
		require.NotEmpty(t, id)
		channelID = id
	})

	if channelID == "" {
		t.Fatal("channel creation failed, skipping lifecycle tests")
	}

	defer func() {
		_ = env.Run("channel", "delete", channelID, "--yes")
	}()

	t.Run("get created channel", func(t *testing.T) {
		result := env.RunSuccess("channel", "get", channelID, "--output", "json")

		var data map[string]interface{}
		env.AssertJSONOutput(result, &data)
	})

	t.Run("list channels includes created", func(t *testing.T) {
		result := env.RunSuccess("channel", "list")

		env.AssertContains(result, "Integration Test Channel")
	})

	t.Run("delete channel", func(t *testing.T) {
		_ = env.RunSuccess("channel", "delete", channelID, "--yes")

		// Verify channel is gone
		getResult := env.Run("channel", "get", channelID)
		require.True(t, getResult.Failed(), "expected get to fail after deletion")
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

// TestLive_RegionList tests the region list command against the live API.
func TestLive_RegionList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default table output", func(t *testing.T) {
		result := env.RunSuccess("region", "list")

		// Should list available regions with name column
		env.AssertContains(result, "NAME")
	})

	t.Run("json output format", func(t *testing.T) {
		result := env.RunSuccess("region", "list", "--output", "json")

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

// TestLive_Whoami tests the whoami command against the live API.
func TestLive_Whoami(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default output", func(t *testing.T) {
		result := env.RunSuccess("whoami")

		// Should show user information
		lines := result.Lines()
		require.NotEmpty(t, lines, "whoami should produce output")
	})

	t.Run("json output", func(t *testing.T) {
		result := env.RunSuccess("whoami", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
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

// TestLive_InvalidProbeGet tests error handling for a non-existent probe.
func TestLive_InvalidProbeGet(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	result := env.RunError("probe", "get", "00000000-0000-0000-0000-000000000000")

	// Should fail with an error message
	combined := result.CombinedOutput()
	require.NotEmpty(t, combined, "expected error output for invalid probe ID")
}

// TestLive_InvalidChannelGet tests error handling for a non-existent channel.
func TestLive_InvalidChannelGet(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	result := env.RunError("channel", "get", "00000000-0000-0000-0000-000000000000")

	combined := result.CombinedOutput()
	require.NotEmpty(t, combined, "expected error output for invalid channel ID")
}

// TestLive_ProbeListOutputFormats tests all output format combinations for probe list.
func TestLive_ProbeListOutputFormats(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	formats := []string{"table", "json", "yaml", "wide"}
	for _, format := range formats {
		t.Run(fmt.Sprintf("format_%s", format), func(t *testing.T) {
			result := env.RunSuccess("probe", "list", "--output", format)

			switch format {
			case "json":
				var data any
				err := json.Unmarshal([]byte(result.Stdout), &data)
				require.NoError(t, err, "expected valid JSON for format %s", format)
			default:
				// Just verify no error for other formats
				_ = result
			}
		})
	}
}

// TestLive_DashboardCommand tests the dashboard overview command.
func TestLive_DashboardCommand(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	result := env.RunSuccess("dashboard")

	// Dashboard should show some kind of overview information
	lines := result.Lines()
	require.NotEmpty(t, lines, "dashboard should produce output")
}

// TestLive_TeamList tests the team list command.
func TestLive_TeamList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default output", func(t *testing.T) {
		result := env.RunSuccess("team", "list")

		// Should show team members - at least the current user
		lines := result.Lines()
		require.NotEmpty(t, lines, "team list should produce output")
	})

	t.Run("json output", func(t *testing.T) {
		result := env.RunSuccess("team", "list", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestLive_StatusPageList tests the status page list command.
func TestLive_StatusPageList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default output", func(t *testing.T) {
		result := env.RunSuccess("status-page", "list")

		// May return 0 status pages
		hasHeader := result.Contains("NAME")
		hasEmptyMessage := result.Contains("No status pages") || result.Contains("no status pages")
		if !hasHeader && !hasEmptyMessage {
			// Some CLIs just print nothing for empty results
			lines := result.Lines()
			_ = lines // Just verify no error
		}
	})

	t.Run("json output", func(t *testing.T) {
		result := env.RunSuccess("status-page", "list", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestLive_APIKeyList tests the API key list command.
func TestLive_APIKeyList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default output", func(t *testing.T) {
		result := env.RunSuccess("apikey", "list")

		// Should show at least one API key (the one being used for auth)
		lines := result.Lines()
		require.NotEmpty(t, lines, "apikey list should produce output")
	})

	t.Run("json output", func(t *testing.T) {
		result := env.RunSuccess("apikey", "list", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestLive_MuteList tests the mute list command.
func TestLive_MuteList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default output", func(t *testing.T) {
		result := env.RunSuccess("mute", "list")

		// May return 0 mutes - verify no error
		_ = result
	})

	t.Run("json output", func(t *testing.T) {
		result := env.RunSuccess("mute", "list", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestLive_LabelList tests the label list command.
func TestLive_LabelList(t *testing.T) {
	env := testutil.NewLiveTestEnv(t)
	defer env.Cleanup()

	t.Run("default output", func(t *testing.T) {
		result := env.RunSuccess("label", "list")

		// May return 0 labels - verify no error
		_ = result
	})

	t.Run("json output", func(t *testing.T) {
		result := env.RunSuccess("label", "list", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}
