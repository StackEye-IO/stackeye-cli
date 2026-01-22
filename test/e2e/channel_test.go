package e2e

import (
	"testing"

	"github.com/StackEye-IO/stackeye-cli/test/e2e/testutil"
)

// TestChannelList tests the channel list command.
func TestChannelList(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("default table output", func(t *testing.T) {
		result := env.RunSuccess("channel", "list")

		// Should contain fixture channel data
		env.AssertContains(result, "Ops Slack")
		env.AssertContains(result, "On-Call Email")
		env.AssertAPICall("GET", "/v1/channels")
	})

	t.Run("json output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("channel", "list", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
		env.AssertAPICall("GET", "/v1/channels")
	})

	t.Run("yaml output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("channel", "list", "--output", "yaml")

		env.AssertContains(result, "name:")
		env.AssertAPICall("GET", "/v1/channels")
	})
}

// TestChannelGet tests the channel get command.
func TestChannelGet(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("get existing channel", func(t *testing.T) {
		result := env.RunSuccess("channel", "get", testutil.ChannelID1)

		env.AssertContains(result, "Ops Slack")
		env.AssertAPICall("GET", "/v1/channels/"+testutil.ChannelID1)
	})

	t.Run("get non-existent channel", func(t *testing.T) {
		result := env.RunError("channel", "get", "99999999-9999-9999-9999-999999999999")

		env.AssertStderrContains(result, "not found")
	})

	t.Run("json output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("channel", "get", testutil.ChannelID1, "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestChannelCreate tests the channel create command.
func TestChannelCreate(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("create email channel", func(t *testing.T) {
		result := env.RunSuccess(
			"channel", "create",
			"--name", "New Email Channel",
			"--type", "email",
			"--email", "alerts@example.com",
		)

		// CLI outputs channel details on success - check for channel name from fixture
		env.AssertContains(result, "New Channel")
		env.AssertAPICall("POST", "/v1/channels")
	})

	t.Run("create slack channel", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess(
			"channel", "create",
			"--name", "New Slack Channel",
			"--type", "slack",
			"--webhook-url", "https://hooks.slack.com/services/test",
		)

		// CLI outputs channel details on success
		env.AssertContains(result, "New Channel")
		env.AssertAPICall("POST", "/v1/channels")
	})

	t.Run("missing required flags", func(t *testing.T) {
		env.ClearAPICalls()
		_ = env.RunError("channel", "create", "--name", "Test")

		// Should fail without type
		env.AssertNoAPICall("POST", "/v1/channels")
	})
}

// TestChannelDelete tests the channel delete command.
func TestChannelDelete(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("delete existing channel", func(t *testing.T) {
		_ = env.RunSuccess("channel", "delete", testutil.ChannelID1, "--yes")

		env.AssertAPICall("DELETE", "/v1/channels/"+testutil.ChannelID1)
	})

	t.Run("delete non-existent channel", func(t *testing.T) {
		result := env.RunError("channel", "delete", "99999999-9999-9999-9999-999999999999", "--yes")

		env.AssertStderrContains(result, "not found")
	})
}

// TestChannelTest tests the channel test command.
func TestChannelTest(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("test existing channel", func(t *testing.T) {
		result := env.RunSuccess("channel", "test", testutil.ChannelID1)

		env.AssertContains(result, "success")
		env.AssertAPICall("POST", "/v1/channels/"+testutil.ChannelID1+"/test")
	})
}
